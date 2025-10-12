downloaderDir = ../downloader
powershell= powershell -Command

build: 
	cd ${downloaderDir} && go build -ldflags="-H=windowsgui" -o ../installer/build/Download-Manager.exe
	${powershell} Copy-Item -Path "${downloaderDir}/pages -Destination" "./build" -Recurse
	${powershell} Copy-Item -Path "../extension -Destination" "./build" -Recurse

clean: 
	if exist build rmdir /S /Q build
	mkdir build