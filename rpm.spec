%define _builddir          .
%define _rpmfilename       %%{NAME}-%%{VERSION}-%%{RELEASE}.%%{ARCH}.rpm
%define _source_payload    w9.lzdio
%define _binary_payload    w9.lzdio
%define _filedir           %(pwd)

Name:    xmlsect
Summary: Quickly query an XML file using XPath 1.0.
Version: 0.4.0
Release: 1
License: MPL-2.0
Group:   Development/Tools
Vendor:  David Gamba
Source0: xmlsect
Source1: xmlsect.1

%description
Quickly query an XML file using XPath 1.0.

%prep
set -x
%setup -n %{buildroot} -c -T

cp %{_sourcedir}/xmlsect .
cp %{_sourcedir}/xmlsect.1 .

%build
set -x
if [ ! -e %{_rpmdir} ]; then
  mkdir -p %{_rpmdir}
fi

mkdir -p %{buildroot}/%{_bindir}
mkdir -p %{buildroot}/%{_mandir}/man1

cp -L %{_sourcedir}/xmlsect   %{buildroot}/%{_bindir}/xmlsect
cp -L %{_sourcedir}/xmlsect.1 %{buildroot}/%{_mandir}/man1/xmlsect.1
gzip %{buildroot}/%{_mandir}/man1/xmlsect.1

rm %{buildroot}/xmlsect
rm %{buildroot}/xmlsect.1

%files
%attr(0755,root,root) %{_bindir}/xmlsect
%attr(0644,root,root) %{_mandir}/man1/xmlsect.1.gz
