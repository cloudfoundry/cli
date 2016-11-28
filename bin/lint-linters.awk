# TODO: REMOVE
# /models\.RouterGroups composite literal uses unkeyed fields \(vet\)/ \
# when https://go-review.googlesource.com/#/c/22318/ is in our released version

!(\
       /should have comment or be unexported.*\(golint\)/ \
    || /should have comment \(or a comment on this block\) or be unexported.*\(golint\)/ \
    || /comment on exported type .+ should be of the form.*\(golint\)/ \
    || /returns unexported type.*\(golint\)/ \
    || /should not use dot imports.*\(golint\)/ \
    || /\"windows\".*\(goconst\)/ \
    || /command\/.*"-1".*\(goconst\)/ \
    || /models\.RouterGroups composite literal uses unkeyed fields \(vet\)/ \
    || /cf\/minimum_api_versions\.go:5:1:warning: _ is unused \(deadcode\)/ \
    || /error return value not checked \(defer .*\) \(errcheck\)/ \
)
