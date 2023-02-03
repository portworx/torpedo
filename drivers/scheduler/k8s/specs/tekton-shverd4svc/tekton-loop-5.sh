while true;
do
for x in {0..20}
do
  cd /Users/apatil/tekton/pipeline-scripts/
   tkn pipeline start "tekton-tests-pipeline" \
  "-w=name=workspace,volumeClaimTemplateFile=${HOME}/tekton/pipeline-scripts/pvc.yaml" \
  "--pod-template=${HOME}/tekton/pipeline-scripts/pod.yaml" \
  --use-param-defaults --namespace tekton-5
done
sleep 900

for x in {0..20}
do
  cd /Users/apatil/tekton/pipeline-scripts/
   tkn pipeline start "tekton-tests-pipeline" \
  "-w=name=workspace,volumeClaimTemplateFile=${HOME}/tekton/pipeline-scripts/pvc.yaml" \
  "--pod-template=${HOME}/tekton/pipeline-scripts/pod.yaml" \
  --use-param-defaults --namespace tekton-5
done
sleep 900

for x in {0..20}
do
  cd /Users/apatil/tekton/pipeline-scripts/
   tkn pipeline start "tekton-tests-pipeline" \
  "-w=name=workspace,volumeClaimTemplateFile=${HOME}/tekton/pipeline-scripts/pvc.yaml" \
  "--pod-template=${HOME}/tekton/pipeline-scripts/pod.yaml" \
  --use-param-defaults --namespace tekton-5
done
sleep 900

for x in {0..20}
do
  cd /Users/apatil/tekton/pipeline-scripts/
   tkn pipeline start "tekton-tests-pipeline" \
  "-w=name=workspace,volumeClaimTemplateFile=${HOME}/tekton/pipeline-scripts/pvc.yaml" \
  "--pod-template=${HOME}/tekton/pipeline-scripts/pod.yaml" \
  --use-param-defaults --namespace tekton-5
done
sleep 900

for x in {0..20}
do
  cd /Users/apatil/tekton/pipeline-scripts/
   tkn pipeline start "tekton-tests-pipeline" \
  "-w=name=workspace,volumeClaimTemplateFile=${HOME}/tekton/pipeline-scripts/pvc.yaml" \
  "--pod-template=${HOME}/tekton/pipeline-scripts/pod.yaml" \
  --use-param-defaults --namespace tekton-5
done
sleep 900
tkn pr delete --all -n tekton-5 -f
sleep 300
tkn pr delete --all -n tekton-5 -f
sleep 300
tkn pr delete --all -n tekton-5 -f

done
