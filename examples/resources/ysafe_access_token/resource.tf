resource "ysafe_access_token" "token" {
    label = "company1"                                              # Identifier for this PIN
    pin = "555555"                                                  # PIN used for sign-in
    expiry = 456                                                    # (Optional) Validity duration in seconds from creation
}