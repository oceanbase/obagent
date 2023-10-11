%define _binaries_in_noarch_packages_terminate_build   0
%define _unpackaged_files_terminate_build              0

# Turn off the brp-python-bytecompile script
%global __os_install_post %(echo '%{__os_install_post}' | sed -e 's!/usr/lib[^[:space:]]*/brp-python-bytecompile[[:space:]].*$!!g')

Name: obagent
Summary: obagent program
Group: alipay/oceanbase
Version:4.2.1
Release: %(echo $RELEASE)%{?dist}
URL: https://github.com/oceanbase/obagent
License: MulanPSL - 2.0
BuildArch: x86_64 aarch64
BuildRoot: %{_tmppath}/%{name}-%{version}-%{release}-root-%(%{__id_u} -n)
Prefix: /home/admin

%description
obagent program

%define _prefix /home/admin

%prep
rm -rf %{_sourcedir}/%{name}
mkdir %{_sourcedir}/%{name}
cp -r $OLDPWD/../*[!rpm]* %_sourcedir/%{name}

%build

export GOROOT=`go env GOROOT`
export GOPATH=`go env GOPATH`
export PATH=$PATH:$GOROOT/bin
export PATH=$PATH:$GOPATH/bin

cd %_sourcedir/%{name}
make pre-build build-release

%install
cd %_sourcedir/%{name}
mkdir -p $RPM_BUILD_ROOT/%{_prefix}/obagent/
mkdir -p $RPM_BUILD_ROOT/%{_prefix}/obagent/bin
mkdir -p $RPM_BUILD_ROOT/%{_prefix}/obagent/log
mkdir -p $RPM_BUILD_ROOT/%{_prefix}/obagent/run
mkdir -p $RPM_BUILD_ROOT/%{_prefix}/obagent/tmp
mkdir -p $RPM_BUILD_ROOT/%{_prefix}/obagent/backup
mkdir -p $RPM_BUILD_ROOT/%{_prefix}/obagent/pkg_store
mkdir -p $RPM_BUILD_ROOT/%{_prefix}/obagent/task_store
mkdir -p $RPM_BUILD_ROOT/%{_prefix}/obagent/position_store
mkdir -p $RPM_BUILD_ROOT/%{_prefix}/obagent/site-packages

cp bin/*                $RPM_BUILD_ROOT/%{_prefix}/obagent/bin
cp -r etc               $RPM_BUILD_ROOT/%{_prefix}/obagent/conf

sed -e 's|${obagent.home.path}|%{_prefix}/obagent|' \
    $RPM_BUILD_ROOT/%{_prefix}/obagent/conf/scripts/obagent.service \
    > $RPM_BUILD_ROOT/obagent.service.tmp
install -p -D -m 0644 $RPM_BUILD_ROOT/obagent.service.tmp \
    %{buildroot}%{_sysconfdir}/systemd/system/multi-user.target.wants/obagent.service
rm $RPM_BUILD_ROOT/obagent.service.tmp

%files
%{_sysconfdir}/systemd/system/multi-user.target.wants/obagent.service
%defattr(755,admin,admin)
%dir %{_prefix}/obagent/
%dir %{_prefix}/obagent/bin
%dir %{_prefix}/obagent/conf
%dir %{_prefix}/obagent/log
%dir %{_prefix}/obagent/run
%dir %{_prefix}/obagent/tmp
%dir %{_prefix}/obagent/backup
%dir %{_prefix}/obagent/pkg_store
%dir %{_prefix}/obagent/task_store
%dir %{_prefix}/obagent/position_store
%{_prefix}/obagent/bin/ob_mgragent
%{_prefix}/obagent/bin/ob_monagent
%{_prefix}/obagent/bin/ob_agentd
%{_prefix}/obagent/bin/ob_agentctl
%{_prefix}/obagent/conf/*.yaml
%config(noreplace) %{_prefix}/obagent/conf/config_properties/*.yaml
%{_prefix}/obagent/conf/module_config/*.yaml
%{_prefix}/obagent/conf/.config_secret.key
%{_prefix}/obagent/conf/shell_templates/*.yaml
%{_prefix}/obagent/conf/prometheus_config/*.yaml
%{_prefix}/obagent/conf/prometheus_config/rules/*.yaml
