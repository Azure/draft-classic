#!groovy

def gitBranch = ''
def gitCommit = ''

def azure = [
  container: 'prow',
  storageAccount: 'deisprow',
  storageKey: '5c0a4a1e-dd9f-4189-90a1-e1a58508394e',
]

def registries = [
  quay: [
    staging: [
      name: 'quay-staging',
      email: 'deisci+jenkins@deis.com',
      username: 'deisci+jenkins',
      credentials: 'c67dc0a1-c8c4-4568-a73d-53ad8530ceeb',
    ],
    production: [
      name: 'quay-production',
      email: 'deis+jenkins@deis.com',
      username: 'deis+jenkins',
      credentials: '8317a529-10f7-40b5-abd4-a42f242f22f0',
    ],
  ]
]

def isMaster = { String branch ->
  branch == "remotes/origin/master"
}

def deriveCommit = {
  commit = sh(returnStdout: true, script: 'git rev-parse HEAD').trim()
  mergeCommitParents = sh(returnStdout: true, script: "echo ${commit} | git log --pretty=%P -n 1 --date-order").trim()

  if (mergeCommitParents.length() > 40) { // is PR
    echo 'More than one merge commit parent signifies that the merge commit is not the actual PR commit'
    echo "Changing commit from '${commit}' to '${mergeCommitParents.take(40)}'"
    commit = mergeCommitParents.take(40)
  }
  commit
}

def dist = { String commit ->
  String version = "git-${commit}"

  sh """
    make build-cross docker-binary
    VERSION=${version} make dist checksum
    VERSION="canary" make dist checksum
  """
}

def dockerBuildAndPush = { Map registry, String commit ->
  String server = registry.name.contains('dockerhub') ? '' : 'quay.io'
  String registryPrefix = registry.name.contains('quay') ? 'quay.io/' : ''
  String imagePrefix = registry.name.contains('staging') ? 'deisci' : 'deis'
  String version = registry.name.contains('staging') ? "git-${commit}" : 'canary'

  sh """
    docker login -e="${registry.email}" -u="${registry.username}" -p="\${REGISTRY_PASSWORD}" ${server}
    REGISTRY=${registryPrefix} IMAGE_PREFIX=${imagePrefix} VERSION=${version} make docker-build docker-push
  """
}

pipeline {
  agent {
    node {
      label 'linux'
    }
  }

  stages {
    stage('Checkout & Git Info') {
      steps {
        checkout scm
        gitBranch = sh(returnStdout: true, script: 'git describe --all').trim()
        gitCommit = deriveCommit()
      }
    }

    stage('Bootstrap') {
      steps {
        sh 'make bootstrap'
      }
    }

    stage('Test') {
      steps {
        sh 'make test'
      }
    }

    stage('Build') {
      steps {
        sh 'make build-cross docker-binary compress-binary'
      }
    }

    stage('Publish Binaries - Azure') {
      environment {
        AZURE_STORAGE_ACCOUNT = azure.storageAccount
        AZURE_STORAGE_KEY = credentials(azure.storageKey)
      }
      steps {
        dist(gitCommit.take(7))
        sh 'az storage blob upload-batch --source _dist/ --destination ${azure.container}'
      }
    }

    stage('Docker Push - Quay.io') {
      def registry = isMaster(gitBranch) ? registries.quay.production : registries.quay.staging
      environment {
        REGISTRY_PASSWORD = credentials(registry.credentials)
      }
      steps {
        dockerBuildAndPush(registry, gitCommit.take(7))
      }
    }
  }
}
