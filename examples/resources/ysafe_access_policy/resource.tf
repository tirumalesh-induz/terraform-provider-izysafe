resource "ysafe_access_policy" "test" {
    name = "company"                # Name of the folder
    max_size = 123                  # (Optional) Max total folder size (in bytes)
    max_file_size = 23              # (Optional) Max size per file (in bytes)
    max_file_versions = 2           # (Optional) Max versions allowed per file
    remove_older_versions = true    # (Optional) Auto-remove oldest version if limit is reached
    default_ttl_for_files = 123456  # (Optional) Duration (in seconds) to retain a file from the time of its latest version upload
}