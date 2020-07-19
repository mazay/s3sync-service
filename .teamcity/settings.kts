import jetbrains.buildServer.configs.kotlin.v2019_2.*
import jetbrains.buildServer.configs.kotlin.v2019_2.buildFeatures.PullRequests
import jetbrains.buildServer.configs.kotlin.v2019_2.buildFeatures.commitStatusPublisher
import jetbrains.buildServer.configs.kotlin.v2019_2.buildFeatures.dockerSupport
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

version = "2020.1"

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

    buildType(UnitTesting)
    buildType(DockerBuild)
    buildType(Build)
    buildType(Release)

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

object UnitTesting : BuildType({
    name = "Unit Testing"

    allowExternalStatus = true

    params {
        param("teamcity.build.default.checkoutDir", "src/s3sync-service")
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
            workingDir = "src"
            name = "Linter check"
            scriptContent = """
                #!/usr/bin/env bash

                ${'$'}{GOBIN}/golint -set_exit_status .
            """.trimIndent()
            formatStderrAsError = true
        }
        script {
            workingDir = "src"
            name = "Go get dependencies"
            scriptContent = "go mod vendor"
            formatStderrAsError = true
        }
        script {
            workingDir = "src"
            name = "Go run unit tests"
            scriptContent = "go test"
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

object DockerBuild : BuildType({
    name = "Docker build"

    allowExternalStatus = true

    params {
        param("teamcity.build.default.checkoutDir", "src/s3sync-service")
        param("env.RELEASE_VERSION", "%teamcity.build.branch%")
        param("reverse.dep.S3syncService_Release.RELEASE_VERSION", "env.RELEASE_VERSION")
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
            name = "Docker multi-arch"
            scriptContent = """
                #!/usr/bin/env bash
                
                if [ -z "${'$'}{RELEASE_VERSION}" ]; then
                    exit 1
                else
                    if [ "${'$'}{RELEASE_VERSION}" = "master" ]; then
                        RELEASE_VERSION="latest"
                    fi
                fi
                
                make docker-multi-arch
            """.trimIndent()
            formatStderrAsError = true
        }
    }


    triggers {
        vcs {
        }
    }

    dependencies {
        snapshot(UnitTesting){
            onDependencyFailure = FailureAction.FAIL_TO_START
        }
    }

    features {
        dockerSupport {
            loginToRegistry = on {
                dockerRegistryId = "PROJECT_EXT_5"
            }
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

object Build : BuildType({
    name = "Build"

    artifactRules = "src/s3sync-service-*"

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
            workingDir = "src"
            name = "Go get dependencies"
            scriptContent = "go mod vendor"
            formatStderrAsError = true
        }
        script {
            workingDir = "src"
            name = "Go build"
            scriptContent = """
                #!/usr/bin/env bash

                os_list=( "darwin" "freebsd" "linux" "windows" )
                arch_list=( "386" "amd64" )

                for os in "${'$'}{os_list[@]}"
                do
                	for arch in "${'$'}{arch_list[@]}"
                  do
                    GOOS=${'$'}{os} GOARCH=${'$'}{arch} go build
                    if [[ ${'$'}{os} == "windows" ]]
                    then
                      filename="s3sync-service.exe"
                    else
                      filename="s3sync-service"
                    fi
                      tar -czvf  s3sync-service-${'$'}{os}-${'$'}{arch}.tar.gz ${'$'}{filename}
                  done
                done
            """.trimIndent()
            formatStderrAsError = true
        }
    }

    dependencies {
        snapshot(UnitTesting){
            onDependencyFailure = FailureAction.FAIL_TO_START
        }
    }
})

object Release : BuildType({
    name = "Release"

    params {
        param("teamcity.build.default.checkoutDir", "src/s3sync-service")
        param("env.RELEASE_VERSION", "")
        param("env.RELEASE_CHANGELOG", "")
        checkbox("env.DRAFT_RELEASE", "true",
                checked = "true", unchecked = "false")
        checkbox("env.PRE_RELEASE", "true",
                checked = "true", unchecked = "false")
        password(
                "env.GITHUB_TOKEN",
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
            name = "Release"
            scriptContent = """
                #!/usr/bin/env bash

                ADDITIONAL_KEYS="-"
                ATTACHMENTS=""

                cat >release.md <<EOF
                ${'$'}{RELEASE_VERSION}

                ${'$'}{RELEASE_CHANGELOG}

                **image:** \`zmazay/s3sync-service:${'$'}{RELEASE_VERSION}\`
                EOF

                if [[ ${'$'}{DRAFT_RELEASE} == true ]]
                then
                  ADDITIONAL_KEYS="${'$'}{ADDITIONAL_KEYS}d"
                fi

                if [[ ${'$'}{PRE_RELEASE} == true ]]
                then
                  ADDITIONAL_KEYS="${'$'}{ADDITIONAL_KEYS}p"
                fi

                if [[ ${'$'}{ADDITIONAL_KEYS} == "-" ]]
                then
                  ADDITIONAL_KEYS=""
                fi

                for artifact in s3sync-service-*
                do
                  ATTACHMENTS="${'$'}{ATTACHMENTS} -a ${'$'}{artifact}"
                done

                hub release create ${'$'}{ADDITIONAL_KEYS} -F release.md ${'$'}{RELEASE_VERSION} ${'$'}{ATTACHMENTS}
            """.trimIndent()
            formatStderrAsError = true
        }
    }

    dependencies {
        dependency(Build) {
            snapshot {
                onDependencyFailure = FailureAction.FAIL_TO_START
            }

            artifacts {
                artifactRules = "s3sync-service-*"
            }
        }
        snapshot(DockerBuild) {
            onDependencyFailure = FailureAction.FAIL_TO_START
        }
    }
})
