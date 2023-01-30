#!groovy

pipeline {
	agent {
		label "porx-builder"
	}

	stages {
		stage("Run Torpedo CBT") {
			steps {
				build(job: "CBT/torpedo-cbt", parameters: [string(name: "TP_GIT_COMMIT", value: GIT_COMMIT)])
			}
		}
	}
}
