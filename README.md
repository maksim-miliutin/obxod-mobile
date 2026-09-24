# obxod-mobile

Bypasses DPI blocking of TLS on Android, with no root and no server in between.
The app runs as a local VpnService and steps into outgoing TLS hellos, so the
inspector cannot read the site's name while the server still reads the request
whole. Traffic goes straight to the sites.

It cannot help where a site is blocked by its address, or where only a list of
allowed sites gets through. Android only.

## Building

The app links a Go library built from `core/`; it is not committed. Rebuild it
with `gomobile` (needs the Android SDK and NDK), or run `core/bind.ps1`:

    cd core
    gomobile bind -target=android -androidapi 24 -o ../app/android/app/libs/obxod.aar ./mobile
