{
  description = "Siren - Wails + Nuxt Application";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs = { self, nixpkgs }: let
    system = "x86_64-linux";
    pkgs = import nixpkgs { inherit system; };
    
    # Agrupamos as dependências do sistema necessárias para compilar a WebView nativa
    linuxDeps = with pkgs; [
      pkg-config
      gtk3
      webkitgtk_4_1 # O motor de renderização da janela no Linux
      cairo
      gdk-pixbuf
      glib
      libsoup_3
    ];

  in {
    # O foco agora é o ambiente de desenvolvimento reprodutível
    devShells.${system}.default = pkgs.mkShell {
      buildInputs = with pkgs; [
        go
        wails
        nodejs_22 # Necessário para rodar o Nuxt (pode trocar para a versão que preferir)
        # pnpm      # Descomente se preferir pnpm ao invés de npm
      ] ++ linuxDeps;
      
      # Variáveis de ambiente essenciais. 
      # O LD_LIBRARY_PATH garante que o binário encontre as libs C em tempo de execução.
      # O XDG_DATA_DIRS previne crashes relacionados a cursores e temas do GTK.
      shellHook = ''
        export LD_LIBRARY_PATH=${pkgs.lib.makeLibraryPath linuxDeps}:$LD_LIBRARY_PATH
        export XDG_DATA_DIRS=${pkgs.gsettings-desktop-schemas}/share/gsettings-schemas/${pkgs.gsettings-desktop-schemas.name}:${pkgs.gtk3}/share/gsettings-schemas/${pkgs.gtk3.name}:$XDG_DATA_DIRS
        
        echo "========================================="
        echo "🧜‍♀️ Bem-vindo ao shell de desenvolvimento do Siren!"
        echo "Go version: $(go version)"
        echo "========================================="
      '';
    };

    # Nota sobre o build final:
    # Empacotar Wails via buildGoModule é um pouco mais complexo porque envolve compilar 
    # o frontend e embutir no Go. Por enquanto, recomendo focar no devShell.
    # Na hora de gerar o binário de produção, o próprio `wails build` dentro deste shell 
    # fará o trabalho pesado gerando o executável final na pasta build/bin.
  };
}