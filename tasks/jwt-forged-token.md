# Requests succeed with tokens the server never issued

## Problem Statement

Users can access accounts they should not. A request with a token the server did not issue is being honored. Only tokens this service signed (and that have not expired) may be accepted.
