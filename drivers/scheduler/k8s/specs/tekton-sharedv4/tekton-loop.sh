#This is an example script to run tkn pipeline in a loop
while true;
 do
   for x in {0..20}
   do
     tkn pipeline start "tekton-tests-pipeline" \
     "-w=name=workspace,volumeClaimTemplateFile=${HOME}/tekton/pipeline-scripts/pvc.yaml" \
     "--pod-template=${HOME}/tekton/pipeline-scripts/pod.yaml" \
     --use-param-defaults --namespace tekton
   done
   sleep 900
 
   for x in {0..20}
   do
     tkn pipeline start "tekton-tests-pipeline" \
     "-w=name=workspace,volumeClaimTemplateFile=${HOME}/tekton/pipeline-scripts/pvc.yaml" \
     "--pod-template=${HOME}/tekton/pipeline-scripts/pod.yaml" \
     --use-param-defaults --namespace tekton
   done
   sleep 900
  
   for x in {0..20}
   do
     tkn pipeline start "tekton-tests-pipeline" \
     "-w=name=workspace,volumeClaimTemplateFile=${HOME}/tekton/pipeline-scripts/pvc.yaml" \
     "--pod-template=${HOME}/tekton/pipeline-scripts/pod.yaml" \
     --use-param-defaults --namespace tekton
   done
   sleep 900

   for x in {0..20}
   do
     tkn pipeline start "tekton-tests-pipeline" \
     "-w=name=workspace,volumeClaimTemplateFile=${HOME}/tekton/pipeline-scripts/pvc.yaml" \
     "--pod-template=${HOME}/tekton/pipeline-scripts/pod.yaml" \
     --use-param-defaults --namespace tekton
   done
   sleep 900

   for x in {0..20}
   do
     tkn pipeline start "tekton-tests-pipeline" \
     "-w=name=workspace,volumeClaimTemplateFile=${HOME}/tekton/pipeline-scripts/pvc.yaml" \
     "--pod-template=${HOME}/tekton/pipeline-scripts/pod.yaml" \
     --use-param-defaults --namespace tekton
   done
   sleep 900

   tkn pr delete --all -n tekton -f
   sleep 300
   tkn pr delete --all -n tekton -f
   sleep 300
   tkn pr delete --all -n tekton -f

done
