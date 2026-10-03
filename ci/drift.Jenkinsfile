// Nightly check that the real AWS infrastructure still matches iac/. Anyone
// who edits a resource in the console makes `terraform plan` see a difference.
// -detailed-exitcode: 0 = identical, 1 = error, 2 = drift.
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
    timeout(time: 20, unit: 'MINUTES')
    disableConcurrentBuilds()
  }

  stages {
    stage('Plan') {
      steps {
        container('terraform') {
          withCredentials([
            usernamePassword(credentialsId: 'aws-tf-readonly',
                             usernameVariable: 'AWS_ACCESS_KEY_ID',
                             passwordVariable: 'AWS_SECRET_ACCESS_KEY'),
            string(credentialsId: 'tf-admin-cidr', variable: 'TF_VAR_admin_cidr'),
          ]) {
            script {
              // Same flags as the PR plan: -lock=false because the read-only
              // user cannot write the state lock, and the same admin_cidr, or
              // every night would report the security group as drifted.
              def code = sh(returnStatus: true, script: '''
                cd iac
                terraform init -input=false
                terraform plan -input=false -lock=false -detailed-exitcode
              ''')
              env.PLAN_EXIT = "${code}"
              if (code == 2) {
                unstable('Terraform drift detected')
              } else if (code != 0) {
                error("terraform plan failed with exit code ${code}")
              }
            }
          }
        }
      }
    }
  }

  post {
    unstable {
      script {
        if (env.PLAN_EXIT == '2') {
          notifyDiscord("⚠️ Terraform drift: the real infrastructure differs from iac/", 16753920)
        }
      }
    }
    failure {
      script { notifyDiscord("❌ Nightly drift check failed to run", 15158332) }
    }
  }
}
