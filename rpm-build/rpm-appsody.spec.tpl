###############################################################################
# Spec file Template (rename to RPM-PACKAGE.spec)
################################################################################
#
Summary: PACKAGE_NAME CLI RPM
Name: PACKAGE_NAME
Version: PACKAGE_VERSION
Release: 1
License: ASL 2.0
URL: http://appsody.dev
Group: System
Packager: Michele Chilanti
Requires: bash
Source0: %{name}
#Source1: CONTROLLER_BASE_URL/%{name}-controller

%description
The PACKAGE_NAME Command-line Interface 

%prep
################################################################################
# Create the build tree and copy the files from the development directories    #
# into the build tree.                                                         #
################################################################################
echo "BUILDROOT = $RPM_BUILD_ROOT"

exit
%build

%install

mkdir -p %{buildroot}/%{_bindir}
# This is not needed since we install everything in /usr/bin
# mkdir -p %{buildroot}/%{_datadir}/.appsody
#
install -m 0755 %{SOURCE0} %{buildroot}/%{_bindir}
#install -m 0755 %{SOURCE1} %{buildroot}/%{_bindir}


%files
%{_bindir}/%{name}
#%{_bindir}/%{name}-controller

%pre

%post
echo "Checking prerequisites..."
docker ps &> /dev/null
if [ $? -eq 0 ]; then
  echo "Done."
else
  echo "[Warning] Docker not detected. Please ensure docker is installed and running before using appsody."
fi

%postun
#rm -rf $HOME/.appsody

%clean
rm -rf %{buildroot}/%{_bindir}
#rm -rf %{buildroot}/%{_datadir}

%changelog
* Thu May 9 2019 NAME HERE <appsodydev@gmail.com>
  - This is the first release of the appsody RPM package.
