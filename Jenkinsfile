#!groovy

def gitBranch = ''
def gitCommit = ''

def azure = [
  container: 'draft',
  storageAccount: 'azuredraft',
  storageKey: '84724ec5-f4a8-4eff-8a66-10db32ea5d3e',
]

def registries = [
  dockerhub: [
    production: [
      name: 'microsoft',
      email: '',
      username: '',
      credentials: '',
    ],
  ]
]

def wrapId = { String envVar, credentialsId ->
  [[
    $class: 'StringBinding',
    credentialsId: credentialsId,
    variable: envVar,
  ]]
}

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
    VERSION=${version} make dist checksum
    VERSION="canary" make dist checksum
  """
}

def dockerBuildAndPush = { Map registry, String commit ->
  String server = 'docker.io'
  String registryPrefix = 'docker.io/'
  String version = registry.name.contains('staging') ? "git-${commit}" : 'canary'

  sh """
    docker login -e="${registry.email}" -u="${registry.username}" -p="\${REGISTRY_PASSWORD}" ${server}
    REGISTRY=${registryPrefix} IMAGE_PREFIX=${registry.name} VERSION=${version} make docker-build docker-push
  """
}

node('linux') {
  env.GOBIN = env.GOPATH + "/bin"
  env.PATH = env.GOBIN + ":" + env.PATH
  def workdir = env.GOPATH + "/src/github.com/Azure/draft"

  dir(workdir) {
    stage('Checkout & Git Info') {
      checkout scm
      gitBranch = sh(returnStdout: true, script: 'git describe --all').trim()
      gitCommit = deriveCommit()
    }

    stage('Bootstrap') {
      sh 'make bootstrap'
    }

    stage('Test') {
      sh 'make test'
    }

    stage('Build') {
      def buildTarget = isMaster(gitBranch) ? 'build-cross' : 'build'

      sh "make ${buildTarget}"
    }

    stage('Publish Binaries - Azure') {
      if (isMaster(gitBranch)) {
        dist(gitCommit.take(7))

        env.AZURE_STORAGE_ACCOUNT = azure.storageAccount
        withCredentials(wrapId('AZURE_STORAGE_KEY', azure.storageKey)) {
          sh "az storage blob upload-batch --source _dist/ --destination ${azure.container} --pattern *.tar.gz*"
        }
      } else {
        echo "git branch not 'master'; skipping binary publishing."
      }
    }

    stage('Docker Push - DockerHub') {
      if (isMaster(gitBranch)) {
        withCredentials(wrapId('REGISTRY_PASSWORD', registries.dockerhub.production.credentials)) {
          dockerBuildAndPush(registries.dockerhub.production, gitCommit.take(7))
        }
      } else {
        echo "git branch not 'master'; skipping dockerhub publishing."
      }
    }
  }
}
