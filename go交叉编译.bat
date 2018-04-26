set CGO_ENABLED=0
set GOROOT_BOOTSTRAP=C:/Go
::x86¿é
set GOARCH=386
set GOOS=windows
call make.bat --no-clean
  
set GOOS=linux
call make.bat --no-clean
  
set GOOS=freebsd
call make.bat --no-clean
  
set GOOS=darwin
call make.bat --no-clean
::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::
  
::x64¿é
set GOARCH=amd64
set GOOS=linux
call make.bat --no-clean
::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::
  
::arm¿é
set GOARCH=arm
set GOOS=linux
call make.bat --no-clean
::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::
  
set GOARCH=386
set GOOS=windows
go get github.com/nsf/gocode
pause