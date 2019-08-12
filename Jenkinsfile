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
                // NOTE: change of this sh call should be in sync with  
                //       './scripts/build.sh' and './codewind-che-sidecar/build.sh'. 
                sh '''#!/usr/bin/env bash
                    echo "Starting build the Codewind Che plugin sidecar container..."
                    set -eu

                    BLUE='\033[1;34m'
                    NC='\033[0m'

                    # Build the sidecar image
                    printf "${BLUE}Building the Codewind sidecar image${NC}\n"
                    cd ./codewind-che-sidecar 

                    # Extract the filewatcherd codebase
                    if [ -d "codewind-filewatchers" ]; then
                        rm -rf codewind-filewatchers
                    fi

                    git clone https://github.com/eclipse/codewind-filewatchers.git

                    docker build -t codewind-che-sidecar .
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
