// Pipeline for the multibranch job: PR builds verify, nothing is pushed.
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
    // Baked into the frontend bundle at build time (see frontend/Dockerfile).
    // Must be the real public URL: an empty value falls back to localhost.
    APP_URL = 'https://app.52-74-96-78.sslip.io'

    // Unfixed CVEs are ignored: nothing can be done about them in this repo,
    // so failing on them would block every PR until upstream ships a patch.
    TRIVY_ARGS = '--severity CRITICAL --ignore-unfixed --exit-code 1 --no-progress'
  }

  stages {
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
              --opt build-arg:NEXT_PUBLIC_API_BASE_URL=$APP_URL \
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
