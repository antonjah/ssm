{ lib, buildGoModule, fetchFromGitHub }:

buildGoModule rec {
  pname = "ssm";
  version = "0.0.1";

  src = fetchFromGitHub {
    owner = "antonjah";
    repo = "ssm";
    rev = "v${version}";
    hash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";
  };

  vendorHash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";

  meta = with lib; {
    description = "ssm - a TUI for managing ssh connections";
    homepage = "https://github.com/antonjah/ssm";
    license = licenses.mit;
    maintainers = [ maintainers.antonjah ];
    mainProgram = "ssm";
  };
}