# GoReleaser copies the prebuilt static binary in; no build stage needed.
# The renderer is pure Go (CGO disabled) with no network use, so scratch is
# enough — no libc, no CA certificates required.
FROM scratch
COPY mermaid /mermaid
ENTRYPOINT ["/mermaid"]
