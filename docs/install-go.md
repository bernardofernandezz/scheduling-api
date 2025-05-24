# Installing Go on Windows

This guide provides step-by-step instructions for installing Go on a Windows system.

## 1. Download the Go Installer

1. Visit the official Go download page at [https://go.dev/dl/](https://go.dev/dl/)
2. Click on the **Microsoft Windows** installer. The download should begin automatically.
   - The filename will look like `go1.20.4.windows-amd64.msi` (version number may differ)
   - Choose the `.msi` installer for Windows.

## 2. Run the Installer

1. Locate the downloaded MSI file in your Downloads folder
2. Double-click on the installer to run it
3. Follow the installation wizard prompts:
   - Accept the license agreement
   - Choose the installation location (default is `C:\Program Files\Go\` or `C:\Go\`)
   - Click "Install" to begin the installation
4. Wait for the installation to complete
5. Click "Finish" when done

## 3. Set Up Environment Variables

The installer should automatically add Go to your PATH environment variable, but it's good to verify this:

1. Open the Start menu and search for "Environment Variables"
2. Click on "Edit the system environment variables"
3. Click the "Environment Variables" button at the bottom
4. In the "System variables" section, find the "Path" variable and click "Edit"
5. Check if `C:\Go\bin` (or your custom installation path) is in the list
   - If not, click "New" and add it manually
6. Click "OK" on all dialog boxes to save changes

Additionally, you may want to set up a GOPATH environment variable:

1. In the "User variables" section, click "New"
2. Set the variable name as `GOPATH`
3. Set the variable value as the path where you want to store Go projects (e.g., `C:\Users\username\go`)
4. Click "OK" to save

## 4. Verify the Installation

1. Open a new PowerShell window (important: open a new window to load the updated environment variables)
2. Run the following command to check if Go is installed correctly:

```powershell
go version
```

You should see output like:
```
go version go1.20.4 windows/amd64
```

3. Verify that the environment is set up correctly:

```powershell
go env
```

This will display all Go environment variables.

## 5. Test with a Simple Program

1. Create a new directory for your test program:

```powershell
mkdir C:\GoTest
cd C:\GoTest
```

2. Create a file named `hello.go` with the following content:

```go
package main

import "fmt"

func main() {
    fmt.Println("Hello, Go!")
}
```

3. Build and run the program:

```powershell
go run hello.go
```

You should see the output: `Hello, Go!`

## 6. Running Your Scheduling API

Now that Go is installed, you can return to your scheduling API project:

```powershell
cd C:\Users\obern\Documents\Developer\playground\scheduling-api\cmd\api
go run main.go
```

If you encounter dependency issues, you may need to download the required packages:

```powershell
go mod download
```

Or if you're not using Go modules yet:

```powershell
go get -u ./...
```

## Troubleshooting

### Command Not Found
If you still get "command not found" after installation, try:
- Reopening your terminal/PowerShell
- Logging out and back in to refresh environment variables
- Rebooting your computer

### Permission Issues
If you encounter permission issues:
- Try running PowerShell as Administrator
- Check if your antivirus is blocking Go

### Module Issues
If you have issues with Go modules:
- Ensure `GO111MODULE` is set correctly: `go env -w GO111MODULE=on`
- Try initializing a new module: `go mod init example.com/myproject`

## Additional Resources

- [Official Go Documentation](https://go.dev/doc/)
- [Go Tour for learning Go](https://go.dev/tour/)
- [Effective Go guide](https://go.dev/doc/effective_go)

