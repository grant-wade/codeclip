pipeline {
    agent {
        docker {
            image 'golang:1.24'
            label 'docker'
        }
    }
    
    environment {
        // You can add additional environment variables if needed
        BUILD_CONFIGURATION = 'Release'
    }
    
    stages {
        stage('Checkout') {
            steps {
                checkout scm
            }
        }
        
        stage('Build') {
            parallel {
                stage('Build for Linux') {
                    steps {
                        script {
                            def binaryName = "codeclip-linux-amd64"
                            echo "Building for linux/amd64"
                            sh """
                                env GOOS=linux GOARCH=amd64 go build -o ${binaryName} main.go
                            """
                        }
                    }
                }
                stage('Build for macOS') {
                    steps {
                        script {
                            def binaryName = "codeclip-darwin-amd64"
                            echo "Building for darwin/amd64"
                            sh """
                                env GOOS=darwin GOARCH=amd64 go build -o ${binaryName} main.go
                            """
                        }
                    }
                }
                stage('Build for Windows') {
                    steps {
                        script {
                            def binaryName = "codeclip-windows-amd64.exe"
                            echo "Building for windows/amd64"
                            sh """
                                env GOOS=windows GOARCH=amd64 go build -o ${binaryName} main.go
                            """
                        }
                    }
                }
            }
        }
        
        stage('Generate Checksums') {
            steps {
                sh """
                    sha256sum codeclip-linux-amd64 > codeclip-linux-amd64.sha256
                    sha256sum codeclip-darwin-amd64 > codeclip-darwin-amd64.sha256
                    sha256sum codeclip-windows-amd64.exe > codeclip-windows-amd64.exe.sha256
                """
            }
        }
        
        stage('Archive Artifacts') {
            steps {
                archiveArtifacts artifacts: 'codeclip-*', allowEmptyArchive: false
            }
        }
    }
    
    post {
        always {
            cleanWs()
        }
    }
}
