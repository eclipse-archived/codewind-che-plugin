#!groovy

pipeline {
    agent {
        label "docker-build"
    }
    
    triggers {	
      issueCommentTrigger('trigger_build')	
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

            // This when clause disables PR build uploads; you may comment this out if you want your build uploaded.
            when {
                beforeAgent true
                not {
                    changeRequest()
                }
            }

            steps {
                withDockerRegistry([url: 'https://index.docker.io/v1/', credentialsId: 'docker.com-bot']) {
                    sh '''#!/usr/bin/env bash
                        if [[ $GIT_BRANCH == "master" ]]; then
                            TAG="latest"
                        else
                            TAG=$GIT_BRANCH
                        fi        

                        # Publish docker images with a filter for branch name
                        # Acceptable branch names: master, start with '<number>.<number>'
                        if [[ $GIT_BRANCH == "master" ]] || [[ $GIT_BRANCH =~ ^([0-9]+\\.[0-9]+) ]]; then
                            echo "Publishing docker images for Eclipse Codewind Che Sidecar..."
                            echo "publish.sh eclipse $TAG"
                            ./scripts/publish.sh eclipse $TAG
                        else
                            echo "Skip publishing docker images for $GIT_BRANCH branch"
                        fi
                    '''
                }
            }
        }
    }
}
