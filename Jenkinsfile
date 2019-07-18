#!groovy​

pipeline {
    agent {
    	label "docker-build"
	}
	
	options {
        timestamps()
        skipStagesAfterUnstable()
    }

    stages {
    	stage('Build Docker image') {
            steps {
				sh '''#!/usr/bin/env bash
					echo "Starting build for Eclipse Codewind Che plugin..."
					sh './scripts/build.sh'
					echo "Publishing docker images for Eclipse Codewind Che plugin..."
					sh 'docker tag codewind-che-sidecar eclipse/codewind-che-sidecar'
					sh 'docker push eclipse/codewind-che-sidecar'
					# sh './scripts/publish.sh eclipse'
				'''
            }
        }
    }
}
