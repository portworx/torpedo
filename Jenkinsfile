#!groovy

pipeline {
	agent {
		label "porx-builder"
	}

	stages {
	    stage("Build Stork") {
            steps {
                build(job: "CBT/torpedo-cbt-build", parameters: [string(name: "GIT_BRANCH", value: GIT_BRANCH)])
            }
        }
		stage("Run Torpedo CBT") {
			steps {
				build(job: "CBT/torpedo-cbt", parameters: [string(name: "GIT_BRANCH", value: GIT_BRANCH)])
			}
		}
	}
}
