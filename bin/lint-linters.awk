!(\
    /should have comment or be unexported.*\(golint\)/ || \
    /should have comment \(or a comment on this block\) or be unexported.*\(golint\)/ || \
    /comment on exported type .+ should be of the form.*\(golint\)/ || \
    /returns unexported type.*\(golint\)/ || \
    /should not use dot imports.*\(golint\)/ || \
    /\"windows\".*\(goconst\)/ \
)