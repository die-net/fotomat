# Go 1.6 and VIPS 8.2.2 aren't available in RPM form on Centos 6 or 7,
# making a proper source RPM difficult.  Instead, this is meant to be run by
# Dockerfile.rpm-centos6 which has built VIPS statically into Fotomat at
# /app/bin/fotomat and just needs to be packaged up.

Name: fotomat
Version: 0.0.0
Release: 1%{dist}
Summary: Fast server for resizing JPEG, PNG, GIF, and WebP images
License: Apache License, Version 2.0
Group: System Environment/Daemons
Source: fotomat
URL: https://github.com/die-net/fotomat
Vendor: Aaron Hopkins
Packager: Aaron Hopkins <tools@die.net>

%description
Fotomat is an extremely fast image resizing proxy, enabling on-the-fly
resizing and cropping of JPEG, PNG, GIF, and WebP images.  Written in Go and
using the fast and flexible VIPS image library, it aims to deliver beautiful
images in the shortest time and at the smallest file size possible.

%prep

%build

%install
install -d -m 755 $RPM_BUILD_ROOT/usr/sbin/
install -s -m 755 /app/bin/fotomat $RPM_BUILD_ROOT/usr/sbin/

%files
/usr/sbin/fotomat
