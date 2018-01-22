#!/usr/bin/env bash
#
# Usage: ./e2e.sh
#
# This script assumes an existing single node k8s cluster with helm and draftd installed. It
# installs all the apps under testdata and checks that they pass or fail, depending on their
# success condition.

TESTDIR="$(pwd)/tests"

cd $TESTDIR
PATH="$(pwd)/../bin:$PATH"

. $TESTDIR/testhelpers.sh

echo "# testing apps that are expected to pass"
pushd testdata/good > /dev/null
for app in */; do
    echo "switching to ${app}"
    pushd "${app}" > /dev/null
    # strip trailing forward slash
    app=${app%/}
    draft up -e nowatch
    echo "checking that ${app} v1 was released"
    revision=$(helm list | grep "${app}" | awk '{print $2}')
    name=$(helm list | grep "${app}" | awk '{print $1}')
    if [[ "$revision" != "1" ]]; then
        fail "Expected REVISION == 1, got $revision"
    fi
	pass
    # deploy the app again and check that the update is seen upstream
    draft up -e nowatch
    echo "checking that ${app} v2 was released"
    revision=$(helm list | grep "${app}" | awk '{print $2}')
    if [[ "$revision" != "2" ]]; then
        fail "Expected REVISION == 2, got $revision"
    fi
	pass

    echo "deleting the helm release for ${app}: ${name}"
    # clean up
    helm delete --purge "${name}"
	pass

    popd > /dev/null
done
popd > /dev/null

echo "# testing apps that are expected to fail"
pushd testdata/bad > /dev/null
for app in */; do
    echo "switching to ${app}"
    pushd "${app}" > /dev/null
    # strip trailing forward slash
    app=${app%/}
    draft up
    echo "checking that ${app} v1 was NOT released"
    release=$(helm list | grep "${app}")
    if [[ "$release" != "" ]]; then
        fail "Expected no release to exist , got $release"
    fi
	pass

    popd > /dev/null
done
popd > /dev/null

# echo "# testing watch and changing files"
# pushd testdata/watch > /dev/null
# for app in */; do
#     echo "switching to ${app}"
#     pushd "${app}" > /dev/null
#     # strip trailing forward slash
#     app=${app%/}
#     # start draft up async
#     draftout=$(mktemp)
#     draftpid=$(draftUpAsync $draftout)

#     # wait for initial sync
#     changes=$(expectChangeAndWaitForSync $draftout)
#     if [[ $changes -ne 0 ]]; then
#         if [[ $changes -eq -1 ]]; then
# 			fail "Expected changes and nothing happened" $draftout $draftpid
# 		fi
#         if [[ $changes -eq -2 ]]; then
# 			fail "Tests timed out" $draftout $draftpid
# 		fi
#     fi

#     # loop over scenarios
#     declare -a filesToClean
#     desiredRevision=2
#     mkdir -p .git/subdir/ # we need a .git directory in some scenario and can't commit it
#     while IFS='' read -r line; do
#         # ignore comments and empty lines
#         [[ "$line" =~ ^# ]] && continue
#         [[ -z "$line" ]] && continue

#         read changeDesired files <<< $line

#         # wait between 2 draft deployments to workaround https://github.com/Azure/draft/issues/79
#         # speedup by only do that if changes are expected
#         [[ $changeDesired == "Y" ]] && sleep 20

#         # modify files
#         for f in $files; do
#             file=${f#rm_}
#             if [ "$f" == "$file" ]; then
#                 # Create or modify file
#                 echo "Modifying $f"
#                 echo "something" >> $f
#             else
#                 # Remove file or directory
#                 echo "Remove $file"
#                 rm -r "$file"
#             fi
#             filesToClean+=("$f")
#         done

#         if [[ $changeDesired == "N" ]]; then
#             if [[ $(hasChanged $draftout) == "true" ]]; then
#                 fail "No rebuild was expected if modifying the following files: $files, but we got some" $draftout $draftpid
#             fi

#         elif [[ $changeDesired == "Y" ]]; then
#             echo "Waiting for build to happen"
#             changes=$(expectChangeAndWaitForSync $draftout)
#             if [[ $changes -ne 0 ]]; then
#                 if [[ $changes -eq -1 ]]; then
# 					fail "Expected changes and nothing happened" $draftout $draftpid
# 				fi
#                 if [[ $changes -eq -2 ]]; then
# 					fail "Tests timed out" $draftout $draftpid
# 				fi
#             fi
#             echo "checking that ${app} v${desiredRevision} was released"
#             revision=$(helm list | grep "${app}" | awk '{print $2}')
#             name=$(helm list | grep "${app}" | awk '{print $1}')
#             if [[ "$revision" != "$desiredRevision" ]]; then
#                 fail "Expected REVISION == $desiredRevision, got $revision" $draftout $draftpid
#             fi
#             desiredRevision=$(( $desiredRevision + 1 ))
#         fi

#     done < scenarios.test

#     pass
#     echo "deleting the helm release for ${app}: ${name}"
#     # clean up
#     $(cleanDraftUpAsync $draftout $draftpid)
#     for f in "${filesToClean[@]}"
#     do
#         rm -f "$f"
#     done
#     rm -r .git
#     helm delete --purge "${name}"
#     pass

#     popd > /dev/null
# done
# popd > /dev/null

echo "e2e tests finished."
