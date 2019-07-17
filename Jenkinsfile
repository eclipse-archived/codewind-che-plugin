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
				'''
            }
        }
    }
}
