# Note import returns content from unreachable hosts

## Problem Statement

Users report that `POST /notes/import?url=...` sometimes returns content from hosts that should be unreachable from the service — including when the supplied URL redirects to one. Import must only fetch legitimate external web resources; it must not be usable to reach the service's own network or infrastructure, whether on the initial request or through a redirect. A normal external URL must still return its content.
