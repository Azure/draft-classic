#!groovy

def gitBranch = ''
def gitCommit = ''

def azure = [
  container: 'draft',
  storageAccount: 'azuredraft',
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
  String server = registry.name.contains('dockerhub') ? '' : 'quay.io'
  String registryPrefix = registry.name.contains('quay') ? 'quay.io/' : ''
  String imagePrefix = registry.name.contains('staging') ? 'deisci' : 'deis'
  String version = registry.name.contains('staging') ? "git-${commit}" : 'canary'

  sh """
    docker login -e="${registry.email}" -u="${registry.username}" -p="\${REGISTRY_PASSWORD}" ${server}
    REGISTRY=${registryPrefix} IMAGE_PREFIX=${imagePrefix} VERSION=${version} make docker-build docker-push
  """
}

node('linux') {
  env.GOBIN = env.GOPATH + "/bin"
  env.PATH = env.GOBIN + ":" + env.PATH
  def workdir = env.GOPATH + "/src/github.com/deis/draft"

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
          sh "az storage blob upload-batch --source _dist/ --destination ${azure.container}"
        }
      } else {
        echo "git branch not 'master'; skipping binary publishing."
      }
    }

    stage('Docker Push - Quay.io') {
      def registry = isMaster(gitBranch) ? registries.quay.production : registries.quay.staging

      withCredentials(wrapId('REGISTRY_PASSWORD', registry.credentials)) {
        dockerBuildAndPush(registry, gitCommit.take(7))
      }
    }
  }
}
