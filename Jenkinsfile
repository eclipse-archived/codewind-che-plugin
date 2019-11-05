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
                    
                    echo JGW test: ${env.BRANCH_NAME}
                    echo post JGW test

                    export INSTALLER_REPO="https://github.com/eclipse/codewind-installer.git"
                    export CW_CLI_BRANCH=master

                        # the command below will echo the head commit if the branch exists, else it just exits
                        if [[ -n \$(git ls-remote --heads \$INSTALLER_REPO ${env.BRANCH_NAME}) ]]; then
                            echo "Will use matching ${env.BRANCH_NAME} branch on \$INSTALLER_REPO"
                            export CW_CLI_BRANCH=${env.BRANCH_NAME}
                        else
                            echo "No matching branch on \$INSTALLER_REPO - using \$CW_CLI_BRANCH branch"
                        fi

                    git clone https://github.com/eclipse/codewind-filewatchers.git

                    # We use --no-cache here because we are consuming cwctl from an external resource (which we should never cache)
                    docker build --no-cache --build-arg CW_CLI_BRANCH=$CW_CLI_BRANCH  -t  codewind-che-sidecar .
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

                            if [[ $GIT_BRANCH =~ ^([0-9]+\\.[0-9]+) ]]; then
                                IFS='.' # set '.' as delimiter
                                read -ra TOKENS <<< "$GIT_BRANCH"    
                                IFS=' ' # reset delimiter
                                export TAG_CUMULATIVE=${TOKENS[0]}.${TOKENS[1]}

                                echo "publish.sh eclipse $TAG_CUMULATIVE"
                                ./scripts/publish.sh eclipse $TAG_CUMULATIVE                               
                            fi
                        else
                            echo "Skip publishing docker images for $GIT_BRANCH branch"
                        fi
                    '''
                }
            }
        }
    }
}
