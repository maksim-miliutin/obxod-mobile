# obxod-mobile

Bypasses DPI blocking of TLS on Android, with no root and no server in between.
The app runs as a local VpnService and steps into outgoing TLS hellos, so the
inspector cannot read the site's name while the server still reads the request
whole. Traffic goes straight to the sites.

It cannot help where a site is blocked by its address, or where only a list of
allowed sites gets through. Android only.
