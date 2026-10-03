// Pipeline for the multibranch job. Every build lints, tests, builds and
// scans; only main continues to push the scanned images and bump the tag.
void notifyDiscord(String title, int color) {
  try {
    container('tools') {
      withCredentials([string(credentialsId: 'discord-webhook', variable: 'DISCORD_URL')]) {
        withEnv(["MSG_TITLE=${title}", "MSG_COLOR=${color}"]) {
          sh '''
            jq -n --arg t "$MSG_TITLE" --arg d "$BUILD_URL" --argjson c "$MSG_COLOR" \
              '{embeds: [{title: $t, description: $d, color: $c}]}' \
              | curl -sS -m 10 -H 'Content-Type: application/json' -d @- "$DISCORD_URL" >/dev/null
          '''
        }
      }
    }
  } catch (err) {
    // A broken webhook must not turn a good build red.
    echo "Discord notification failed: ${err}"
  }
}

pipeline {
  agent {
    kubernetes {
      yamlFile 'ci/agent-pod.yaml'
      defaultContainer 'tools'
    }
  }

  options {
    timeout(time: 45, unit: 'MINUTES')
    disableConcurrentBuilds()
  }

  environment {
    AWS_REGION = 'ap-southeast-1'
    ECR_REGISTRY = '334177992720.dkr.ecr.ap-southeast-1.amazonaws.com'

    // Unfixed CVEs are ignored: nothing can be done about them in this repo,
    // so failing on them would block every PR until upstream ships a patch.
    TRIVY_ARGS = '--severity CRITICAL --ignore-unfixed --exit-code 1 --no-progress'
  }

  stages {
    // The tag-bump commit pushed by this pipeline lands on main and would
    // trigger another build; this deletes that build instead of looping.
    stage('Skip bot commits') {
      steps {
        scmSkip(deleteBuild: true, skipPattern: '.*\\[skip ci\\].*')
      }
    }

    stage('Prepare') {
      steps {
        sh 'git config --global --add safe.directory "*"'
        script {
          env.IMAGE_TAG = sh(returnStdout: true, script: 'git rev-parse --short=8 HEAD').trim()
        }
        // bimg is cgo over libvips, so vet/lint/test all need the headers.
        container('golang') {
          sh '''
            apt-get update -qq
            apt-get install -y -qq --no-install-recommends libvips-dev pkg-config
            curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/v1.64.8/install.sh \
              | sh -s -- -b /usr/local/bin v1.64.8
          '''
        }
      }
    }

    stage('Lint') {
      parallel {
        stage('backend') {
          steps {
            container('golang') {
              sh '''
                cd backend
                test -z "$(gofmt -l .)"
                go vet ./...
                golangci-lint run ./...
              '''
            }
          }
        }
        stage('frontend') {
          steps {
            container('node') {
              sh '''
                cd frontend
                npm ci --no-audit --no-fund
                npx next lint
                npx tsc --noEmit
              '''
            }
          }
        }
      }
    }

    // The frontend has no tests yet.
    stage('Test') {
      steps {
        container('golang') {
          sh 'cd backend && go test -race -cover ./...'
        }
      }
    }

    stage('Build images') {
      steps {
        container('buildkit') {
          sh '''
            mkdir -p build
            buildctl-daemonless.sh build \
              --frontend dockerfile.v0 \
              --local context=backend --local dockerfile=backend \
              --output type=docker,name=thai-gov-processor-backend:$IMAGE_TAG,dest=build/backend.tar
            buildctl-daemonless.sh build \
              --frontend dockerfile.v0 \
              --local context=frontend --local dockerfile=frontend \
              --output type=docker,name=thai-gov-processor-frontend:$IMAGE_TAG,dest=build/frontend.tar
          '''
        }
      }
    }

    stage('Trivy gate') {
      steps {
        container('trivy') {
          sh '''
            trivy fs --scanners vuln,secret --skip-dirs build,frontend/node_modules,iac/.terraform $TRIVY_ARGS .
            trivy image --input build/backend.tar $TRIVY_ARGS
            trivy image --input build/frontend.tar $TRIVY_ARGS
          '''
        }
      }
    }

    stage('Terraform plan') {
      when {
        allOf {
          changeRequest()
          expression { sh(returnStatus: true, script: 'git diff --quiet origin/$CHANGE_TARGET...HEAD -- iac/') != 0 }
        }
      }
      steps {
        container('terraform') {
          withCredentials([
            usernamePassword(credentialsId: 'aws-tf-readonly',
                             usernameVariable: 'AWS_ACCESS_KEY_ID',
                             passwordVariable: 'AWS_SECRET_ACCESS_KEY'),
            string(credentialsId: 'tf-admin-cidr', variable: 'TF_VAR_admin_cidr'),
          ]) {
            // -lock=false: the read-only user cannot write the state lock file.
            sh '''
              cd iac
              terraform fmt -check -recursive
              terraform init -input=false
              terraform validate
              terraform plan -input=false -lock=false
            '''
          }
        }
      }
    }

    stage('Push to ECR') {
      when { branch 'main' }
      steps {
        // The node's IAM role supplies the credentials; crane reads the
        // Docker-style config written here, so no access key is stored.
        container('aws-cli') {
          sh '''
            set +x  # the token must not be echoed into the build log
            mkdir -p .docker
            token=$(aws ecr get-login-password --region "$AWS_REGION")
            auth=$(printf 'AWS:%s' "$token" | base64 -w0)
            printf '{"auths":{"%s":{"auth":"%s"}}}' "$ECR_REGISTRY" "$auth" > .docker/config.json
          '''
        }
        container('crane') {
          // Pushes the exact tar Trivy scanned. Tags are IMMUTABLE, so a
          // re-run of the same commit keeps what is already there.
          sh '''
            export DOCKER_CONFIG="$PWD/.docker"
            for name in backend frontend; do
              ref="$ECR_REGISTRY/thai-gov-processor-$name:$IMAGE_TAG"
              if crane digest "$ref" >/dev/null 2>&1; then
                echo "$ref already exists, skipping push"
              else
                crane push "build/$name.tar" "$ref"
              fi
            done
          '''
        }
      }
    }

    stage('Bump image tag') {
      when { branch 'main' }
      steps {
        withCredentials([usernamePassword(credentialsId: 'github-token',
                                          usernameVariable: 'GIT_USER',
                                          passwordVariable: 'GIT_TOKEN')]) {
          sh '''
            export GIT_ASKPASS="$PWD/ci/git-askpass.sh" GIT_TERMINAL_PROMPT=0

            # Another merge may have landed while this build ran.
            git fetch origin main
            git checkout -B main origin/main

            overlay=k8s/overlays/prod/kustomization.yaml
            sed -i "s/newTag: .*/newTag: $IMAGE_TAG/" "$overlay"
            test "$(grep -c "newTag: $IMAGE_TAG" "$overlay")" = 2

            git config user.name jenkins-ci
            git config user.email jenkins-ci@users.noreply.github.com
            git add "$overlay"
            git commit -m "deploy $IMAGE_TAG [skip ci]"
            git push origin main
          '''
        }
      }
    }
  }

  post {
    success {
      script {
        if (env.BRANCH_NAME == 'main') {
          notifyDiscord("✅ ${env.JOB_NAME} #${env.BUILD_NUMBER}: ${env.IMAGE_TAG} pushed, tag bumped", 3066993)
        }
      }
    }
    failure {
      script {
        if (env.BRANCH_NAME == 'main') {
          notifyDiscord("❌ ${env.JOB_NAME} #${env.BUILD_NUMBER} failed", 15158332)
        }
      }
    }
  }
}
