FROM golang:1.8

RUN apt-get update -y
RUN apt-get install -y wget git ncurses-dev time

WORKDIR /tmp
RUN git clone https://github.com/vim/vim.git && cd vim && make && make install

WORKDIR /go/src
COPY .vim /root/.vim
COPY .vimrc /root/.vimrc
COPY ./Godeps /go/src

RUN wget https://raw.githubusercontent.com/pote/gpm/v1.4.0/bin/gpm && chmod +x gpm && mv gpm /usr/local/bin
RUN gpm install
RUN cp -r ./github.com /github.com  # backup packages to root to prevent volume mount removing it

# Install Go binaries that are utilised by the vim-go plugin:
# https://github.com/fatih/vim-go/blob/master/plugin/go.vim#L9
#
# We don't manually install them, we let vim-go handle that
# We use vim's `execute` command to pipe commands
# This helps avoid "Press ENTER or type command to continue"
RUN time vim -c "execute 'silent GoUpdateBinaries' | execute 'quit'"
