{
  description = "Siren - Wails + Nuxt Application";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs = { self, nixpkgs }: let
    # Lista de sistemas que o Flake vai suportar. 
    # Cobre o seu desktop Linux e o MacBook (Intel ou Apple Silicon)
    supportedSystems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
    
    # Função auxiliar para gerar saídas dinâmicas para cada arquitetura
    forAllSystems = nixpkgs.lib.genAttrs supportedSystems;
  in {
    
    devShells = forAllSystems (system: let
      pkgs = import nixpkgs { inherit system; };

      # Dependências comuns para todos os ambientes
      commonDeps = with pkgs; [
        go
        wails
        nodejs_22 
      ];

      # Dependências exclusivas do Linux (GTK, WebKit, CGO)
      linuxDeps = with pkgs; [
        pkg-config
        gtk3
        webkitgtk_4_1
        libayatana-appindicator
        libdbusmenu-gtk3
        cairo
        gdk-pixbuf
        glib
        libsoup_3
        gst_all_1.gstreamer
        gst_all_1.gst-plugins-base
        gst_all_1.gst-plugins-good
      ];

    in {
      default = pkgs.mkShell {
        # O Nix injeta as dependências certas testando o SO base
        buildInputs = commonDeps 
          ++ pkgs.lib.optionals pkgs.stdenv.isLinux linuxDeps;
        
        # O shellHook também precisa ser condicional.
        # O macOS não precisa (nem deve) exportar variáveis do GTK.
        shellHook = ''
          ${pkgs.lib.optionalString pkgs.stdenv.isLinux ''
            export LD_LIBRARY_PATH=${pkgs.lib.makeLibraryPath linuxDeps}:$LD_LIBRARY_PATH
            export XDG_DATA_DIRS=${pkgs.gsettings-desktop-schemas}/share/gsettings-schemas/${pkgs.gsettings-desktop-schemas.name}:${pkgs.gtk3}/share/gsettings-schemas/${pkgs.gtk3.name}:$XDG_DATA_DIRS
          ''}
          
          echo "========================================="
          echo "🧜‍♀️ Bem-vindo ao shell de desenvolvimento do Siren!"
          echo "Plataforma: ${if pkgs.stdenv.isDarwin then "macOS" else "Linux"}"
          echo "Go version: $(go version)"
          echo "========================================="
        '';
      };
    });
  };
}