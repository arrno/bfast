{
  "version": "VERSION",
  "homepage": "https://blazingly.fast",
  "description": "CLI badger for registering repos with blazingly.fast and inserting the badge",
  "license": "Proprietary",
  "architecture": {
    "64bit": {
      "url": "https://github.com/arrno/bfast/releases/download/TAG/bfast_VERSION_windows_amd64.zip",
      "hash": "SHA256_WINDOWS_AMD64"
    }
  },
  "bin": [
    "bfast.exe"
  ],
  "checkver": {
    "url": "https://api.github.com/repos/arrno/bfast/releases/latest",
    "jsonpath": "$.tag_name"
  },
  "autoupdate": {
    "architecture": {
      "64bit": {
        "url": "https://github.com/arrno/bfast/releases/download/$version/bfast_$version_windows_amd64.zip"
      }
    }
  }
}
