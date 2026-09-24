# Build the Go core into the Android library the app links.
# Run from core/; needs gomobile with the Android SDK and NDK on PATH.
$ErrorActionPreference = 'Stop'
New-Item -ItemType Directory -Force -Path '../app/android/app/libs' | Out-Null
gomobile bind -target=android -androidapi 24 -o ../app/android/app/libs/obxod.aar ./mobile
