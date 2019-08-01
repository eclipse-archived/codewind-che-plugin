#!groovy

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
                    ./scripts/build.sh
                '''
            }
        }
        
        stage('Publish Docker image') {
            steps {
                withDockerRegistry([url: 'https://index.docker.io/v1/', credentialsId: 'docker.com-bot']) {
                    sh '''#!/usr/bin/env bash
                        if [[ $GIT_BRANCH == "master" ]]; then
                            TAG="latest"
                        else
                            TAG=$GIT_BRANCH
                        fi        

                        if [ -z $CHANGE_ID ]; then
                            echo "Publishing docker images for Eclipse Codewind Che Sidecar..."
                            echo "publish.sh eclipse $TAG"
                            ./scripts/publish.sh eclipse $TAG
                        else
                            echo "Skip publishing docker images for the PR build"
                        fi
                    '''
                }
            }
        }
    }
}
