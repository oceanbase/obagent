#!/bin/bash
cd `git rev-parse --show-toplevel`

sed_agent_spec() {
  sed "s/Version:.*/Version:$1/" rpm/obagent.spec >rpm/obagent.spec.bak &&
    mv -f rpm/obagent.spec.bak rpm/obagent.spec
}

sed_agent_makefile() {
  sed "s/VERSION=.*/VERSION=$1/" Makefile.common > Makefile.common.bak &&
    mv -f Makefile.common.bak Makefile.common
}

change_obagent_version(){
  OBAGENT_FORMAL_VERSION=`echo $OBAGENT_VERSION | awk -F - '{print $1}'` # e.g. 3.3.0
  echo $OBAGENT_FORMAL_VERSION > rpm/obagent-version.txt
  sed_agent_spec $OBAGENT_FORMAL_VERSION
  sed_agent_makefile $OBAGENT_VERSION
}

case X$1 in
    Xset-obagent-version)
  OBAGENT_VERSION="$2"
  echo "set-obagent-version $OBAGENT_VERSION"
  change_obagent_version
        ;;
    Xshow-version)
	cat rpm/obagent-version.txt
	;;
    *)
	echo "Usage: change_version.sh set-obagent-version|show-version"
	echo "Examples:"
	echo "       change_version.sh set-obagent-version 4.1.0"
	echo "       change_version.sh show-version"
esac
