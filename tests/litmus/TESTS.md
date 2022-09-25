| TEST DESCRIPTION                 | TEST                             | EXPECTATION                              |
|----------------------------------|----------------------------------|------------------------------------------|
| No Data loss, No Data corruption | https://github.com/portworx/torpedo/tree/master/tests/litmus   Tl;dr:   https://github.com/portworx/torpedo/blob/master/tests/litmus/README.md#running-litmus-in-docker                                                                                                                                 | The target file should not be corrupt.   |
| Resync                           | Create a volume with a replication factor of 3.,Call the nodes A, B and C.  Write 1GB of data to the volume from Node A.  Immediately turn nodes B off.  Write 1 more GB of data on node A.  Now turn node A off.  Use the volume on node C and verify that you can read all 2GB of data correctly. | The contents of the data must be intact. |

