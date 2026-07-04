# File download can escape the attachments directory

## Problem Statement

The attachment download endpoint can return files from outside the attachments directory when the name contains path segments. It must only ever serve files inside the attachments directory.
