// Pipeline for the multibranch job. Branch flow: feature -> dev -> staging -> main.
//   PR (feature)      lint, test, build, scan (+ terraform plan when iac/ changes)
//   PR dev->staging   verify the promoted image exists in ECR (no rebuild)
//   PR staging->main  same
//   dev               full CI, push the scanned images, bump the staging overlay tag
//   staging           verify only: the tag already arrived via the merge
//   main              promote: copy the staging tag into the prod overlay
// Images are built once on dev; staging and prod only ever move that tag.
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

void ecrLogin() {
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
}

// Commits the new tag to `branch` as the bot. [skip ci] stops the build that
// the push would otherwise trigger.
void bumpTag(String branch, String overlay, String tag) {
  withCredentials([usernamePassword(credentialsId: 'github-token',
                                    usernameVariable: 'GIT_USER',
                                    passwordVariable: 'GIT_TOKEN')]) {
    withEnv(["BUMP_BRANCH=${branch}", "BUMP_FILE=${overlay}", "BUMP_TAG=${tag}"]) {
      sh '''
        export GIT_ASKPASS="$PWD/ci/git-askpass.sh" GIT_TERMINAL_PROMPT=0

        # Another merge may have landed while this build ran.
        git fetch origin "$BUMP_BRANCH"
        git checkout -B "$BUMP_BRANCH" "origin/$BUMP_BRANCH"

        sed -i "s/newTag: .*/newTag: $BUMP_TAG/" "$BUMP_FILE"
        test "$(grep -c "newTag: $BUMP_TAG" "$BUMP_FILE")" = 2
        if git diff --quiet -- "$BUMP_FILE"; then
          echo "$BUMP_FILE is already at $BUMP_TAG"
          exit 0
        fi

        git config user.name jenkins-ci
        git config user.email jenkins-ci@users.noreply.github.com
        git add "$BUMP_FILE"
        git commit -m "deploy $BUMP_TAG [skip ci]"
        git push origin "$BUMP_BRANCH"
      '''
    }
  }
}

boolean runs(List modes) {
  return modes.contains(env.MODE)
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
    STAGING_OVERLAY = 'k8s/overlays/staging/kustomization.yaml'
    PROD_OVERLAY = 'k8s/overlays/prod/kustomization.yaml'

    // Unfixed CVEs are ignored: nothing can be done about them in this repo,
    // so failing on them would block every PR until upstream ships a patch.
    TRIVY_ARGS = '--severity CRITICAL --ignore-unfixed --exit-code 1 --no-progress'
  }

  stages {
    // Decides what this build does (env.MODE): skip, ci, build, verify or promote.
    stage('Classify') {
      steps {
        sh 'git config --global --add safe.directory "*"'
        script {
          env.IMAGE_TAG = sh(returnStdout: true, script: 'git rev-parse --short=8 HEAD').trim()

          // Read the top commit directly. The scmSkip plugin looks at the
          // changelog instead, which after several merges in a row contains
          // feature commits, so a bot commit slipped through and built.
          // Not for PRs: the head of a promotion PR is often the bot's own
          // `deploy ... [skip ci]` commit and must still report a status.
          def message = sh(returnStdout: true, script: 'git log -1 --pretty=%B').trim()
          def bot = !env.CHANGE_ID && message.contains('[skip ci]')

          def target = env.CHANGE_TARGET ?: ''
          def source = env.CHANGE_BRANCH ?: ''
          def promotion = (target == 'staging' && source == 'dev') || (target == 'main' && source == 'staging')

          // Only PRs can be judged docs-only. A queued branch build checks out the
          // branch head when it starts, so after several merges in a row
          // `HEAD^1 HEAD` shows only the last merge and could hide code from an
          // earlier one. The PR diff covers the whole PR. Anything that cannot be
          // diffed counts as "not docs only".
          def docsOnly = false
          if (env.CHANGE_ID) {
            docsOnly = sh(returnStdout: true, script: '''
              files=$(git diff --name-only origin/$CHANGE_TARGET...HEAD 2>/dev/null || true)
              other=$(printf '%s\\n' "$files" | grep -vcE '^(docs/.*|[^/]*\\.md)$' || true)
              if [ -n "$files" ] && [ "$other" = 0 ]; then
                echo true
              else
                echo false
              fi
            ''').trim() == 'true'
          }

          if (bot) {
            env.MODE = 'skip'
          } else if (env.CHANGE_ID) {
            env.MODE = promotion ? 'verify' : (docsOnly ? 'skip' : 'ci')
          } else if (env.BRANCH_NAME == 'dev') {
            env.MODE = 'build'
          } else if (env.BRANCH_NAME == 'staging') {
            env.MODE = 'verify'
          } else if (env.BRANCH_NAME == 'main') {
            env.MODE = 'promote'
          } else {
            env.MODE = 'skip'
          }
          echo "MODE=${env.MODE} (bot=${bot}, docsOnly=${docsOnly}, promotion=${promotion})"
        }
      }
    }

    stage('CI') {
      when { expression { runs(['ci', 'build']) } }
      stages {
        stage('Prepare') {
          steps {
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
      }
    }

    stage('Push to ECR') {
      when { expression { runs(['build']) } }
      steps {
        ecrLogin()
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

    stage('Bump staging tag') {
      when { expression { runs(['build']) } }
      steps {
        script { bumpTag('dev', env.STAGING_OVERLAY, env.IMAGE_TAG) }
      }
    }

    // The tag being promoted is the one in the staging overlay. Both images
    // must exist in ECR, otherwise Argo CD would deploy a tag that cannot be pulled.
    stage('Verify promoted image') {
      when { expression { runs(['verify', 'promote']) } }
      steps {
        ecrLogin()
        container('crane') {
          sh '''
            export DOCKER_CONFIG="$PWD/.docker"
            tags=$(sed -n 's/^ *newTag: //p' "$STAGING_OVERLAY" | sort -u)
            [ -n "$tags" ] || { echo "no newTag in $STAGING_OVERLAY"; exit 1; }
            test "$(echo "$tags" | wc -l)" = 1 || { echo "staging overlay has different tags: $tags"; exit 1; }
            for name in backend frontend; do
              crane digest "$ECR_REGISTRY/thai-gov-processor-$name:$tags" >/dev/null \
                || { echo "image $name:$tags is not in ECR"; exit 1; }
            done
            echo "$tags" > .promote-tag
          '''
        }
        script { env.PROMOTE_TAG = readFile('.promote-tag').trim() }
      }
    }

    // Only when this merge changed the staging tag. Running on every main build
    // would roll a `git revert` of a prod bump forward again on the next merge.
    stage('Promote to prod') {
      when { expression { runs(['promote']) } }
      steps {
        script {
          def changed = sh(returnStdout: true,
            script: 'git diff --name-only HEAD^1 HEAD -- "$STAGING_OVERLAY" 2>/dev/null || true').trim()
          if (changed) {
            bumpTag('main', env.PROD_OVERLAY, env.PROMOTE_TAG)
            env.PROMOTED = 'true'
          } else {
            echo 'The staging tag did not change in this merge, nothing to promote.'
          }
        }
      }
    }
  }

  post {
    success {
      script {
        if (env.MODE == 'build') {
          notifyDiscord("✅ ${env.JOB_NAME} #${env.BUILD_NUMBER}: ${env.IMAGE_TAG} pushed, staging tag bumped", 3066993)
        } else if (env.PROMOTED == 'true') {
          notifyDiscord("✅ ${env.JOB_NAME} #${env.BUILD_NUMBER}: ${env.PROMOTE_TAG} promoted to prod", 3066993)
        }
      }
    }
    failure {
      script {
        if (!env.CHANGE_ID) {
          notifyDiscord("❌ ${env.JOB_NAME} #${env.BUILD_NUMBER} failed", 15158332)
        }
      }
    }
  }
}
