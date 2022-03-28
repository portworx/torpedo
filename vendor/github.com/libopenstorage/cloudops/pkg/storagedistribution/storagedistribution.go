package storagedistribution

import (
	"fmt"
	"math"
	"sort"

	"github.com/libopenstorage/cloudops"
	"github.com/libopenstorage/cloudops/pkg/utils"
	"github.com/libopenstorage/openstorage/api"
	"github.com/sirupsen/logrus"
)

/*
   ====================================
   Storage Distribution Algorithm
   ====================================

  The storage distribution algorithm provides an optimum
  storage distribution strategy for a given set of following inputs:
  - Requested IOPS from cloud storage.
  - Minimum capacity for the whole cluster.
  - Number of zones in the cluster.
  - Number of instances in the cluster.
  - A storage decision matrix.

  TODO:
   - Take into account instance types and their supported drives
   - Take into account the effect on the overall throughput when multiple drives are attached
     on the same instance.
*/

// GetStorageUpdateConfig returns the storage configuration for updating
// an instance's storage based on the requested new capacity.
// To meet the new capacity requirements this function with either:
// - Resize existing disks
// - Add more disks
// This is based of the ResizeOperationType input argument. If no such input is
// provided then this function tries Resize first and then an Add.
// The algorithms for Resize and Add are explained with their respective function
// definitions.
func GetStorageUpdateConfig(
	request *cloudops.StoragePoolUpdateRequest,
	decisionMatrix *cloudops.StorageDecisionMatrix,
) (*cloudops.StoragePoolUpdateResponse, *cloudops.StorageDecisionMatrixRow, error) {
	logUpdateRequest(request)

	switch request.ResizeOperationType {
	case api.SdkStoragePool_RESIZE_TYPE_ADD_DISK:
		// Add drives equivalent to newDeltaCapacity
		return AddDisk(request, decisionMatrix)
	case api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK:
		// Resize existing drives equivalent to newDeltaCapacity
		return ResizeDisk(request, decisionMatrix)
	default:
		// Auto-mode. Try resize first then add
		resp, row, err := ResizeDisk(request, decisionMatrix)
		if err != nil {
			return AddDisk(request, decisionMatrix)
		}
		return resp, row, err
	}
}

// AddDisk tries to satisfy the StoragePoolUpdateRequest by adding more disks
// to the existing storage pool. Following is a high level algorithm/steps used
// to achieve this:
// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// - Calculate deltaCapacity = input.RequestedCapacity - input.CurrentCapacity					 //
// - Calculate currentDriveSize from the request.								 //
// - Calculate the requiredDriveCount for achieving the deltaCapacity.						 //
// - Find out if any rows from the decision matrix fit in our new configuration					 //
//      - Filter out the rows which do not have the same input.DriveType					 //
//      - Filter out rows which do not fit input.CurrentDriveSize in row.MinSize and row.MaxSize		 //
//      - Filter out rows which do not fit requiredDriveCount in row.InstanceMinDrives and row.InstanceMaxDrives //
// - Pick the 1st row from the decision matrix as your candidate.						 //
// - If no row found:												 //
//     - failed to AddDisk											 //
// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func AddDisk(
	request *cloudops.StoragePoolUpdateRequest,
	decisionMatrix *cloudops.StorageDecisionMatrix,
) (*cloudops.StoragePoolUpdateResponse, *cloudops.StorageDecisionMatrixRow, error) {
	if err := validateUpdateRequest(request); err != nil {
		return nil, nil, err
	}

	currentCapacity := request.CurrentDriveCount * request.CurrentDriveSize
	deltaCapacity := request.DesiredCapacity - currentCapacity

	logrus.Debugf("check if we can add drive(s) for atleast: %d GiB", deltaCapacity)
	dm := utils.CopyDecisionMatrix(decisionMatrix)

	currentDriveSize := request.CurrentDriveSize
	if currentDriveSize == 0 {
		// No drives have been provisioned yet.
		// Lets select a row witch matches the deltaCapacity
		// TODO: Need to start with different increments here. (Vsphere: TestCase: 8)
		currentDriveSize = deltaCapacity
	}

	// Calculate the driveCount required to fit the deltaCapacity
	requiredDriveCount := uint64(math.Ceil(float64(deltaCapacity) / float64(currentDriveSize)))

	updatedTotalDrivesOnNodes := requiredDriveCount + request.TotalDrivesOnNode

	// Filter the decision matrix and check if there any rows which satisfy our requirements.
	dm = dm.FilterByDriveType(request.CurrentDriveType)
	if len(dm.Rows) == 0 {
		return nil, nil, &cloudops.ErrStorageDistributionCandidateNotFound{
			Reason: fmt.Sprintf("found no candidates which have current drive type: %s", request.CurrentDriveType),
		}
	}

	driveRanges := make(map[string]string)
	driveRangeStr := make([]string, 0)
	for _, row := range dm.Rows {
		driveRanges[fmt.Sprintf("[%d GiB -> %d GiB (%s)]", row.MinSize, row.MaxSize, row.DriveType)] = ""
	}

	for k := range driveRanges {
		driveRangeStr = append(driveRangeStr, k)
	}

	dm = dm.FilterByDriveSizeRange(currentDriveSize)
	if len(dm.Rows) == 0 {
		sort.Strings(driveRangeStr)
		return nil, nil, &cloudops.ErrStorageDistributionCandidateNotFound{
			Reason: fmt.Sprintf("found no candidates for adding a new disk of existing size: %d GiB. "+
				"Only drives in following size ranges are supported: %v", currentDriveSize, driveRangeStr),
		}
	}

	maxDriveCount := dm.Rows[0].InstanceMaxDrives
	for _, row := range dm.Rows {
		if row.InstanceMaxDrives > maxDriveCount {
			maxDriveCount = row.InstanceMaxDrives
		}
	}

	dm = dm.FilterByDriveCount(updatedTotalDrivesOnNodes)
	if len(dm.Rows) == 0 {
		return nil, nil, &cloudops.ErrStorageDistributionCandidateNotFound{
			Reason: fmt.Sprintf("node has reached it's maximum supported drive count: %d", maxDriveCount),
		}
	}

	row := dm.Rows[0]
	printCandidates("AddDisk Candidate", []cloudops.StorageDecisionMatrixRow{row}, 0, 0)

	instStorage := &cloudops.StoragePoolSpec{
		DriveType:        row.DriveType,
		DriveCapacityGiB: currentDriveSize,
		DriveCount:       uint64(requiredDriveCount),
	}
	prettyPrintStoragePoolSpec(instStorage, "AddDisk")
	resp := &cloudops.StoragePoolUpdateResponse{
		InstanceStorage:     []*cloudops.StoragePoolSpec{instStorage},
		ResizeOperationType: api.SdkStoragePool_RESIZE_TYPE_ADD_DISK,
	}
	return resp, &row, nil
}

// ResizeDisk tries to satisfy the StoragePoolUpdateRequest by expanding existing disks
// from the storage pool. Following is a high level algorithm/steps used
// to achieve this:
// ////////////////////////////////////////////////////////////////////////////////////////////////
// - Calculate deltaCapacity = input.RequestedCapacity - input.CurrentCapacity		        //
// - Calculate deltaCapacityPerDrive = deltaCapacityPerNode / input.CurrentNumberOfDrivesInPool //
// - Filter out the rows which do not have the same input.DriveType			        //
// - Filter out the rows which do not have the same input.IOPS				        //
// - Filter out rows which do not fit input.CurrentDriveSize in row.MinSize and row.MaxSize     //
// - Sort the rows by IOPS								        //
// - First row in the filtered decision matrix is our best candidate.			        //
// - If input.CurrentDriveSize + deltaCapacityPerDrive > row.MaxSize:			        //
//       - failed to expand								        //
//   Else										        //
//       - success									        //
// ////////////////////////////////////////////////////////////////////////////////////////////////
func ResizeDisk(
	request *cloudops.StoragePoolUpdateRequest,
	decisionMatrix *cloudops.StorageDecisionMatrix,
) (*cloudops.StoragePoolUpdateResponse, *cloudops.StorageDecisionMatrixRow, error) {
	if err := validateUpdateRequest(request); err != nil {
		return nil, nil, err
	}

	if request.CurrentDriveCount == 0 {
		return nil, nil, &cloudops.ErrInvalidStoragePoolUpdateRequest{
			Request: request,
			Reason: fmt.Sprintf("requested resize operation type cannot be " +
				"accomplished as no existing drives were provided"),
		}
	}

	deltaCapacityPerDrive := calculateDriveCapacity(request)
	logrus.Debugf("Delta capacity per drive: %v", deltaCapacityPerDrive)

	dm := utils.CopyDecisionMatrix(decisionMatrix)

	// We select the first matching row of the matrix as it satisfies the following:
	// 1. same drive type
	// 2. drive size lies between row's min and max size
	// 3. row's IOPS is closest to the current IOPS
	// Filter the decision matrix
	dm = dm.FilterByDriveType(request.CurrentDriveType)
	if len(dm.Rows) == 0 {
		return nil, nil, &cloudops.ErrStorageDistributionCandidateNotFound{
			Reason: fmt.Sprintf("found no candidates which have current drive type: %s", request.CurrentDriveType),
		}
	}

	dm = dm.FilterByIOPS(request.CurrentIOPS)
	if len(dm.Rows) == 0 {
		return nil, nil, &cloudops.ErrStorageDistributionCandidateNotFound{
			Reason: fmt.Sprintf("found no drive candidates that match current IOPS of drives on node: %d", request.CurrentIOPS),
		}
	}

	dm = dm.FilterByDriveSize(request.CurrentDriveSize).SortByIOPS()
	if len(dm.Rows) == 0 {
		return nil, nil, &cloudops.ErrStorageDistributionCandidateNotFound{
			Reason: fmt.Sprintf("found no drive candidates that match current drive size: %d", request.CurrentDriveSize),
		}
	}

	for _, row := range dm.Rows {
		printCandidates("ResizeDisk Candidate", []cloudops.StorageDecisionMatrixRow{row}, 0, 0)
		if request.CurrentDriveSize+deltaCapacityPerDrive > row.MaxSize {
			continue
		}

		instStorage := &cloudops.StoragePoolSpec{
			DriveType:        row.DriveType,
			DriveCapacityGiB: request.CurrentDriveSize + deltaCapacityPerDrive,
			DriveCount:       request.CurrentDriveCount,
		}
		prettyPrintStoragePoolSpec(instStorage, "ResizeDisk")
		resp := &cloudops.StoragePoolUpdateResponse{
			InstanceStorage:     []*cloudops.StoragePoolSpec{instStorage},
			ResizeOperationType: api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK,
		}
		return resp, &row, nil
	}
	return nil, nil, &cloudops.ErrStorageDistributionCandidateNotFound{}
}

func calculateDriveCapacity(request *cloudops.StoragePoolUpdateRequest) uint64 {
	currentCapacity := request.CurrentDriveCount * request.CurrentDriveSize
	deltaCapacity := request.DesiredCapacity - currentCapacity
	// need to round up to avoid cases like:
	// deltaCapacity==5
	// CurrentDriveCount == 2
	// deltaCapacityPerDrive == 2 < DesiredCapacity
	return uint64(math.Ceil(float64(deltaCapacity) / float64(request.CurrentDriveCount)))
}

// GetStorageDistributionForPool tries to determine a drive configuration
// to satisfy the input storage pool requirements. Following is a high level algorithm/steps used
// to achieve this:
//
// ////////////////////////////////////////////////////////////////////////////
// - Calculate minCapacityPerZone = input.MinCapacity / zoneCount	    //
// - Calculate maxCapacityPerZone = input.MaxCapacity / zoneCount	    //
// - Filter the decision matrix based of our requirements:		    //
//     - Filter out the rows which do not have the same input.DriveType	    //
//     - Filter out the rows which do not meet input.IOPS		    //
//     - Sort the decision matrix by IOPS				    //
//     - Sort the decision matrix by Priority				    //
// - instancesPerZone = input.RequestedInstancesPerZone			    //
// - (row_loop) For each of the filtered row:				    //
//     - (instances_per_zone_loop) For instancesPerZone > 0:		    //
//         - Find capacityPerNode = minCapacityPerZone / instancesPerZone   //
//             - (drive_count_loop) For driveCount > row.InstanceMinDrives: //
//                 - driveSize = capacityPerNode / driveCount		    //
//                 - If driveSize within row.MinSize and row.MaxSize:	    //
//                     break drive_count_loop (Found candidate)		    //
//             - If (drive_count_loop) fails/exhausts:			    //
//                   - reduce instancesPerZone by 1			    //
//                   - goto (instances_per_zone_loop)			    //
//               Else found candidate					    //
//                   - break instances_per_zone_loop (Found candidate)	    //
//     - If (instances_per_zone_loop) fails:				    //
//         - Try the next filtered row					    //
//         - goto (row_loop)						    //
// - If (row_loop) fails:						    //
//       - failed to get a candidate					    //
// ////////////////////////////////////////////////////////////////////////////
func GetStorageDistributionForPool(
	decisionMatrix *cloudops.StorageDecisionMatrix,
	request *cloudops.StorageSpec,
	requestedInstancesPerZone uint64,
	zoneCount uint64,
) (*cloudops.StoragePoolSpec, uint64, *cloudops.StorageDecisionMatrixRow, error) {
	logDistributionRequest(request, requestedInstancesPerZone, zoneCount)

	if zoneCount <= 0 {
		return nil, 0, nil, cloudops.ErrNumOfZonesCannotBeZero
	}

	// Filter the decision matrix rows based on the input request
	dm := utils.CopyDecisionMatrix(decisionMatrix)
	dm.FilterByDriveType(request.DriveType).
		FilterByIOPS(request.IOPS).
		SortByIOPS().
		SortByPriority()

	// Calculate min and max capacity per zone
	minCapacityPerZone := request.MinCapacity / uint64(zoneCount)
	maxCapacityPerZone := request.MaxCapacity / uint64(zoneCount)
	var (
		capacityPerNode, instancesPerZone, driveCount, driveSize uint64
		row                                                      cloudops.StorageDecisionMatrixRow
		rowIndex                                                 uint64
	)

row_loop:
	for rowIndex = uint64(0); rowIndex < uint64(len(dm.Rows)); rowIndex++ {
		row = dm.Rows[rowIndex]
		// Favour maximum instances per zone
	instances_per_zone_loop:
		for instancesPerZone = requestedInstancesPerZone; instancesPerZone > 0; instancesPerZone-- {
			capacityPerNode = minCapacityPerZone / uint64(instancesPerZone)
			printCandidates("Candidate", []cloudops.StorageDecisionMatrixRow{row}, instancesPerZone, capacityPerNode)
			// Favour maximum drive count
			// drive_count_loop:
			foundCandidate := false
			for driveCount = row.InstanceMaxDrives; driveCount >= row.InstanceMinDrives; driveCount-- {
				driveSize = capacityPerNode / driveCount
				if driveSize >= row.MinSize && driveSize <= row.MaxSize {
					// Found a candidate
					foundCandidate = true
					break
				}
				if driveCount == row.InstanceMinDrives {
					// We have exhausted the drive_count_loop
					if driveSize < row.MinSize {
						// If the last calculated driveSize is less than row.MinSize
						// that indicates none of the driveSizes in the drive_count_loop
						// were greater than row.MinSize. Lets try with row.MinSize
						driveSize = row.MinSize
						driveCount = row.InstanceMinDrives
						if driveSize*instancesPerZone < maxCapacityPerZone {
							// Found a candidate
							foundCandidate = true
							break
						}
					}
				}
			}

			if !foundCandidate {
				// drive_count_loop failed
				continue instances_per_zone_loop
			}
			break instances_per_zone_loop
		}

		if instancesPerZone == 0 {
			// instances_per_zone_loop failed
			continue row_loop
		}
		// break row_loop
		break row_loop
	}

	if int(rowIndex) == len(dm.Rows) {
		// row_loop failed
		return nil, 0, nil, &cloudops.ErrStorageDistributionCandidateNotFound{}
	}

	// optimize instances per zone
	var optimizedInstancesPerZone uint64
	for optimizedInstancesPerZone = uint64(1); optimizedInstancesPerZone < instancesPerZone; optimizedInstancesPerZone++ {
		// Check if we can satisfy the minCapacityPerZone for this optimizedInstancesPerZone, driveCount and driveSize
		if minCapacityPerZone > optimizedInstancesPerZone*driveCount*driveSize {
			// we are not satisfying the minCapacityPerZone
			continue
		}
		break
	}
	instStorage := &cloudops.StoragePoolSpec{
		DriveType:        row.DriveType,
		DriveCapacityGiB: driveSize,
		DriveCount:       driveCount,
	}
	prettyPrintStoragePoolSpec(instStorage, "getStorageDistributionCandidate returning")
	return instStorage, optimizedInstancesPerZone, &row, nil

}

// validateUpdateRequest validates the StoragePoolUpdateRequest
func validateUpdateRequest(
	request *cloudops.StoragePoolUpdateRequest,
) error {
	currentCapacity := request.CurrentDriveCount * request.CurrentDriveSize
	if currentCapacity > request.DesiredCapacity {
		return &cloudops.ErrCurrentCapacityHigherThanDesired{
			Current: currentCapacity,
			Desired: request.DesiredCapacity,
		}
	}

	newDeltaCapacity := request.DesiredCapacity - currentCapacity
	if newDeltaCapacity == 0 {
		return cloudops.ErrCurrentCapacitySameAsDesired
	}

	if request.CurrentDriveCount > 0 && len(request.CurrentDriveType) == 0 {
		return &cloudops.ErrInvalidStoragePoolUpdateRequest{
			Request: request,
			Reason: fmt.Sprintf("for storage update operation, current drive" +
				"type is required to be provided if drives already exist"),
		}
	}
	return nil
}

func prettyPrintStoragePoolSpec(spec *cloudops.StoragePoolSpec, prefix string) {
	logrus.Infof("%s instStorage: %d X %d GiB %s drives", prefix, spec.DriveCount,
		spec.DriveCapacityGiB, spec.DriveType)
}

func printCandidates(
	msg string,
	candidates []cloudops.StorageDecisionMatrixRow,
	instancePerZone uint64,
	capacityPerNode uint64,
) {
	for _, candidate := range candidates {
		logrus.WithFields(logrus.Fields{
			"MinIOPS":           candidate.MinIOPS,
			"MaxIOPS":           candidate.MaxIOPS,
			"MinSize":           candidate.MinSize,
			"MaxSize":           candidate.MaxSize,
			"DriveType":         candidate.DriveType,
			"Priority":          candidate.Priority,
			"InstanceMinDrives": candidate.InstanceMinDrives,
			"InstanceMaxDrives": candidate.InstanceMaxDrives,
		}).Debugf("%v for %v instances per zone with a total capacity per node %v",
			msg, instancePerZone, capacityPerNode)
	}
}

func logDistributionRequest(
	request *cloudops.StorageSpec,
	requestedInstancesPerZone uint64,
	zoneCount uint64,
) {
	logrus.WithFields(logrus.Fields{
		"IOPS":             request.IOPS,
		"MinCapacity":      request.MinCapacity,
		"InstancesPerZone": requestedInstancesPerZone,
		"ZoneCount":        zoneCount,
	}).Debugf("-- Storage Distribution Pool Request --")
}

func logUpdateRequest(
	request *cloudops.StoragePoolUpdateRequest,
) {
	logrus.WithFields(logrus.Fields{
		"MinCapacity":   request.DesiredCapacity,
		"OperationType": request.ResizeOperationType,
	}).Debugf("-- Storage Distribution Pool Update Request --")
}
