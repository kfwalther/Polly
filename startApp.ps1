# Powershell script to easily kick off both frontend and backend portions in separate Powershell windows.

# Kick off the backend (Go server).
invoke-expression 'cmd /c start powershell -Command {D: ; cd "D:\workspace\Polly\backend"; .\go-server.exe}'
# Kick off the frontend (React app).
invoke-expression 'cmd /c start powershell -Command {D: ; cd "D:\workspace\Polly\ui"; npm start}'
