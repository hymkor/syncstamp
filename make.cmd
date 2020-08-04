setlocal
call :"%1"
endlocal
exit /b

:""
    go build -ldflags "-s -w"
    exit /b

:"386"
    set GOARCH=386
    call :""
    exit /b
