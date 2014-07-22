shell-proxy
===========

proxy server to execute shell commands


## How to execute

### 1. Packaging

	mvn package

### 2. Starting the server

	cd deploy
	java -jar shell-proxy.jar [port]

### 3. Executing commnds by URL

	http://localhost:8080/?command=[command]

### 4. Stopping the server by URL

	http://localhost:8080/stop
