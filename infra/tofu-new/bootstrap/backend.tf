terraform {
  backend "s3" {
    bucket              = "ezvacss-xyz-tofu-state-471575131234-eu-north-1"
    key                 = "ezvacss.xyz/bootstrap/terraform.tfstate"
    region              = "eu-north-1"
    allowed_account_ids = ["471575131234"]
    encrypt             = true
    use_lockfile        = true
  }
}