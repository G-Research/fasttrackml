[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12

$ProgressPreference = 'SilentlyContinue'
$release = Invoke-RestMethod -Method Get -Uri "https://api.github.com/repos/jgiannuzzi/fasttrackml/releases/latest"
$asset = $release.assets | Where-Object name -like *windows_x86_64.zip
$destdir = "$home\bin"
$zipfile = "$env:TEMP\$( $asset.name )"
$zipfilename = [System.IO.Path]::GetFileNameWithoutExtension("$zipfile")

Write-Output "Downloading: $( $asset.name )"
Invoke-RestMethod -Method Get -Uri $asset.browser_download_url -OutFile $zipfile

# Check if an older version of xh.exe exists in '$destdir', if yes, then delete it, if not then download latest zip to extract from
$fmlPath = "${destdir}\fml.exe"
if (Test-Path -Path $fmlPath -PathType Leaf)
{
    Write-Output "Removing previous installation of fml from $destdir"
    Remove-Item -r -fo $fmlPath
}

# Create dir for result of extraction
New-Item -ItemType Directory -Path $destdir -Force | Out-Null

# Decompress the zip file to the destination directory
Add-Type -Assembly System.IO.Compression.FileSystem
$zip = [IO.Compression.ZipFile]::OpenRead($zipfile)
$entries = $zip.Entries | Where-Object { $_.FullName -like '*.exe' }
$entries | ForEach-Object { [IO.Compression.ZipFileExtensions]::ExtractToFile($_, $destdir + "\" + $_.Name) }

# Free the zipfile
$zip.Dispose()
Remove-Item -Path $zipfile

$fmlVersion = & "$destdir\fml.exe" --version

# Inform user where the executables have been put
Write-Output "fml v$( $fmlVersion ) has been installed to:`n - $fmlPath`n"

# Make sure destdir is in the path
$userPath = [System.Environment]::GetEnvironmentVariable('Path', [System.EnvironmentVariableTarget]::User)
$machinePath = [System.Environment]::GetEnvironmentVariable('Path', [System.EnvironmentVariableTarget]::Machine)

# If userPath AND machinePath both do not contain bin, then add it to user path
if (!($userPath.ToLower().Contains($destdir.ToLower())) -and !($machinePath.ToLower().Contains($destdir.ToLower())))
{
    # Update userPath
    $userPath = $userPath.Trim(";") + ";$destdir"

    # Modify PATH for new windows
    Write-Output "`nAdding $destdir directory to the PATH variable."
    [System.Environment]::SetEnvironmentVariable('Path', $userPath, [System.EnvironmentVariableTarget]::User)

    # Modify PATH for current terminal
    Write-Output "`nRefreshing current terminal's PATH for you."
    $Env:Path = $Env:Path.Trim(";") + ";$destdir"

    # Instruct how to modify PATH for other open terminals
    Write-Output "`nFor other terminals, restart them (or the entire IDE if they're within one).`n"

}