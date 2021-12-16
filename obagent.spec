Name: obagent
Summary: ob agent program
Group: alipay/oceanbase
Version: 1.1.0
Release: %(echo $RELEASE)%{?dist}
URL: http://rpm.corp.taobao.com/find.php?q=obagent
License: Commercial
BuildArch: x86_64 aarch64
BuildRoot: %{_tmppath}/%{name}-%{version}-%{release}-root-%(%{__id_u} -n)

%description
obagent program

%define _prefix /home/admin

%prep
rm -rf %{_sourcedir}/%{name}
mkdir -p %_sourcedir/%{name}
cp -r $OLDPWD/* %_sourcedir/%{name}

%build

cd %_sourcedir/%{name}
make build

%install
cd %_sourcedir/%{name}
mkdir -p $RPM_BUILD_ROOT/%{_prefix}/obagent/
mkdir -p $RPM_BUILD_ROOT/%{_prefix}/obagent/bin
mkdir -p $RPM_BUILD_ROOT/%{_prefix}/obagent/conf
mkdir -p $RPM_BUILD_ROOT/%{_prefix}/obagent/log
mkdir -p $RPM_BUILD_ROOT/%{_prefix}/obagent/run
mkdir -p $RPM_BUILD_ROOT/%{_prefix}/obagent/tmp
mkdir -p $RPM_BUILD_ROOT/%{_prefix}/obagent/backup
mkdir -p $RPM_BUILD_ROOT/%{_prefix}/obagent/pkg_store
mkdir -p $RPM_BUILD_ROOT/%{_prefix}/obagent/task_store

cp bin/*                $RPM_BUILD_ROOT/%{_prefix}/obagent/bin
cp -r etc/*             $RPM_BUILD_ROOT/%{_prefix}/obagent/conf

%files
%defattr(755,admin,admin)
%dir %{_prefix}/obagent/
%dir %{_prefix}/obagent/bin
%dir %{_prefix}/obagent/conf
%dir %{_prefix}/obagent/run
%{_prefix}/obagent/bin/monagent
%{_prefix}/obagent/conf/*.yaml
%{_prefix}/obagent/conf/config_properties/*.yaml
%{_prefix}/obagent/conf/module_config/*.yaml
%{_prefix}/obagent/conf/prometheus_config/*.yaml
%{_prefix}/obagent/conf/prometheus_config/rules/*.yaml
