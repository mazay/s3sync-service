import jetbrains.buildServer.configs.kotlin.v2019_2.*
import jetbrains.buildServer.configs.kotlin.v2019_2.buildFeatures.PullRequests
import jetbrains.buildServer.configs.kotlin.v2019_2.buildFeatures.commitStatusPublisher
import jetbrains.buildServer.configs.kotlin.v2019_2.buildFeatures.golang
import jetbrains.buildServer.configs.kotlin.v2019_2.buildFeatures.pullRequests
import jetbrains.buildServer.configs.kotlin.v2019_2.buildSteps.script
import jetbrains.buildServer.configs.kotlin.v2019_2.triggers.vcs
import jetbrains.buildServer.configs.kotlin.v2019_2.vcs.GitVcsRoot

/*
The settings script is an entry point for defining a TeamCity
project hierarchy. The script should contain a single call to the
project() function with a Project instance or an init function as
an argument.

VcsRoots, BuildTypes, Templates, and subprojects can be
registered inside the project using the vcsRoot(), buildType(),
template(), and subProject() methods respectively.

To debug settings scripts in command-line, run the

    mvnDebug org.jetbrains.teamcity:teamcity-configs-maven-plugin:generate

command and attach your debugger to the port 8000.

To debug in IntelliJ Idea, open the 'Maven Projects' tool window (View
-> Tool Windows -> Maven Projects), find the generate task node
(Plugins -> teamcity-configs -> teamcity-configs:generate), the
'Debug' option is available in the context menu for the task.
*/

version = "2019.2"

object GitGithubComMazayS3syncServiceGit : GitVcsRoot({
    name = "git@github.com:mazay/s3sync-service.git"
    url = "git@github.com:mazay/s3sync-service.git"
    branchSpec = "+:refs/heads/*"
    authMethod = uploadedKey {
        userName = "git"
        uploadedKey = "teamcity_github_s3sync_service"
    }
})

project {

    vcsRoot(GitGithubComMazayS3syncServiceGit)

    buildType(Build)

    features {
        feature {
            id = "PROJECT_EXT_2"
            type = "IssueTracker"
            param("secure:password", "")
            param("name", "mazay/s3sync-service")
            param("pattern", """#(\d+)""")
            param("authType", "anonymous")
            param("repository", "https://github.com/mazay/s3sync-service")
            param("type", "GithubIssues")
            param("secure:accessToken", "")
            param("username", "")
        }
    }
}

object Build : BuildType({
    name = "Build"

    allowExternalStatus = true
    artifactRules = "s3sync-service"

    params {
        param("teamcity.build.default.checkoutDir", "src/s3sync-service")
        param("env.DEBIAN_FRONTEND", "noninteractive")
        param("env.GOFLAGS", "-json")
        param("env.GOPATH", "/opt/buildagent/work")
        password(
          "s3sync-service.github.token",
          "credentialsJSON:38d0338a-0796-4eaa-a625-d9b720d9af17",
          label = "Github Token",
          display = ParameterDisplay.HIDDEN,
          readOnly = true
        )
    }

    vcs {
        root(DslContext.settingsRoot)
    }

    steps {
        script {
            name = "Go get dependencies"
            scriptContent = "go mod vendor"
            formatStderrAsError = true
        }
        script {
            name = "Go test"
            scriptContent = "go test"
            formatStderrAsError = true
        }
        script {
            name = "Go build"
            scriptContent = "go build"
            formatStderrAsError = true
        }
    }

    triggers {
        vcs {
        }
    }

    features {
        pullRequests {
            vcsRootExtId = "${DslContext.settingsRoot.id}"
            provider = github {
                authType = token {
                    token = "credentialsJSON:8c15f79d-8a9d-4ab0-9057-7f7bc00883c3"
                }
                filterAuthorRole = PullRequests.GitHubRoleFilter.MEMBER
            }
        }
        golang {
            testFormat = "json"
        }
        commitStatusPublisher {
            vcsRootExtId = "${DslContext.settingsRoot.id}"
            publisher = github {
                githubUrl = "https://api.github.com"
                authType = personalToken {
                    token = "credentialsJSON:8c15f79d-8a9d-4ab0-9057-7f7bc00883c3"
                }
            }
        }
    }
})
