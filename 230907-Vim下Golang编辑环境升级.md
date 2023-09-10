之前`vim`下配置`Golang`开发环境使用的是关于`nvim`搭配`git`子模块实现的。经过一段时间的折腾，探索了一些新的实践，在这里记录一下。

## 使用原生vim
`nvim`等`vim`的分支项目的出现，已经驱使`vim`变得更好。现在`vim8+`已经是满足了我的需要。有兴趣的可以关注下`vim`的发展。主要常用的`vim`插件也是同时支持`nvim`和`vim`的，这样我们无需重新安装其他的`vim`再通过`alise vim`的方式替换掉本机的`vim`了，本质还是嫌麻烦，坚持少就是多。

- `vim`配置文件默认在`~/.vimrc`中，其次可以在`~/.vim/vimrc`下。我选择放在`~/.vim/vimrc`下。
- `vim`的插件安装位置，默认在`~/.vim`下。

首先先在终端输入`vim`查看版本，如果`vim8+`就不需要额外更新，否则可以：

```shell
~ brew install vim
```

## 安装vim-plug
`git`子模块虽然不需要增加额外的插件管理，但是使用起来还是有些不太舒服，比如卸载子模块，更新子模块，没有那么的自然。所以这里切换到了`vim-plug`。

安装`vim-plug`：

```shell
# 安装vim-plug插件管理，到~/.vim/autoload下
~ curl -fLo ~/.vim/autoload/plug.vim --create-dirs \
    https://raw.githubusercontent.com/junegunn/vim-plug/master/plug.vim
```

## 使用vim-plug管理插件
```text
call plug#begin("~/.vim/plugged")
Plug 'fatih/vim-go', { 'do' : ':GoUpdateBinaries' }
Plug 'preservim/nerdtree'

Plug 'junegunn/fzf', { 'dir': '~/.fzf', 'do': './install --all' }
Plug 'junegunn/fzf.vim'

Plug 'lifepillar/vim-solarized8'

Plug 'airblade/vim-rooter'
call plug#end()
```

## 代码自动提醒增强
在`vimrc`中新增自动命令行注册，作用到`go`文件，键入`.`后，会自动执行`Ctrl+x`和`Ctrl+o`唤出代码提醒。

```bash
autocmd filetype go inoremap <buffer> . .<C-x><C-o>
```

## fzf搜索插件增强
虽然经过折腾，这里的fzf已经可以实现在当前项目下，实现全文搜索。但是对于`go`项目来说还需要从依赖的`mod`项目中寻找关键字，这里就不满足了，所以准备扩展下当前fzf的能力。

### 准备如下的shell脚本
- 文件命名为`gf`标识为更广泛范围内搜索的意思。
- 移动脚本到`~/.vim`文件下，并增加可执行权限`chmod +x gf`
- 在`.zshrc`中，新增命令重命名`alias gf='source ~/.vim/gf'`
- 在`go mod`项目下，执行`gf`后，键入`f keyword`即可实现在所有依赖的项目包下，实现关键字搜索。

```shell
#!/bin/sh

modlist=$(go list -m all)
tmp=$(echo "$modlist" | sed 's/ /@/g')

f(){
    local keyword="$1"

    echo "$tmp" | while IFS= read -r line; do
        if test -d "$GOPATH/pkg/mod/$line"; then
            /usr/local/bin/rg --column --line-number --no-heading --color=always -w -i -t go "$keyword" "$GOPATH/pkg/mod/$line" | grep "$keyword"
        else
            # echo "目录不存在"
        fi
    done
}
```


## vimrc配置文件如下
```bash
" 通用
set autowrite " 自动保存
set laststatus=2 " 始终显示状态栏
set statusline=%F\ %l,%c " 控制状态栏显示文件名，行号和列号
autocmd filetype go inoremap <buffer> . .<C-x><C-o>
" 显示行号
nnoremap <leader>sn :set number!<CR>
set autoindent " 设置自动缩进
set confirm " 在处理未保存或只读文件时，弹出确认
set history=1000 " 设置历史记录步数
set tabstop=4 " 设置制表符为4空格
set shiftwidth=4 " 设置自动缩进长度为4空格
set expandtab " vim自动将输入的制表符替换为缩进的空格数
autocmd BufWritePre *.go :silent! GoFmt " 保存文件时自动fmt go代码

" 主题配色
syntax enable
set background=dark
let g:solarized_termtrans=1 " This gets rid of the grey background
colorscheme solarized8
hi SpecialKey ctermbg=None ctermfg=66

" NerdTree
" 显示或隐藏目录
nnoremap <leader>nn :NERDTreeToggle<CR>
" 为某个文件或者文件夹加书签，D删除书签
nnoremap <leader>nb :NERDTreeFromBookmark<CR>
" 打开目录，并定位到当前文件
nnoremap <leader>nf :NERDTreeFind<CR>

" fzf and vim-rooter
" 当前项目内全局查找关键字
nnoremap <leader>f :RG<cr>
" nnoremap <leader>gf :GF<cr>
" 查看最近打开的文件和打开的缓冲区
nnoremap <leader>r :History<cr>
" 查看打开的缓冲区
nnoremap <leader>j :Buffers<cr>
" 当前项目内查找文件
nnoremap <leader>k :Files<cr>
" 当前项目下查找Tag
nnoremap <leader>l :Tags<cr>

" vim-go
" 双击鼠标左键，跳转到光标所在代码定义处
nnoremap <2-LeftMouse> :GoDef <CR>
" 鼠标右键，从代码定义跳转回上一个位置
nnoremap <RightMouse> :GoDefPop <CR>
" Leader+gr 列出调用当前光标下标识符的代码位置
nnoremap <leader>gr :GoCallers <CR>
" Leader+ge 列出当前光标下标识符调用的代码位置
nnoremap <leader>ge :GoCallees <CR>
" 鼠标滚轮向上滚动，会执行向上滚动，类似于按下Ctrl+Y
nnoremap <ScrollWheelUp> <C-Y>
" 鼠标滚轮向下滚动时，会执行向下滚动，类似于Ctrl+E
nnoremap <ScrollWheelDown> <C-E>
" 按下Leader+gd，会执行Go代码跳转，类似于双击鼠标左键
nnoremap <leader>gd :GoDef <CR>
" 按下Leader+gp，会执行Go代码定义跳转为上一个位置，类似于鼠标右键
nnoremap <leader>gp :GoDefPop <CR>
" 按下Shift+K时，会显示当前光标下标识符文档注释
nnoremap <S-K> :GoDoc<cr>
" 按下Shift+M时，会显示当前光标下标识符的详细信息
nnoremap <S-M> :GoInfo<cr>
" 按下Shift+T时，会跳转到当前光标下标识符的类型定义
nnoremap <S-T> :GoDefType<cr>
" 按下Shift+L时，会执行Go结构体标签（Tag）的添加操作
"nnoremap <S-L> :GoAddTag<cr>
nnoremap <S-L> :GoIfErr <CR>
" 按下Shift+P时，会列出当前光标下接口（interface）的所有实现
nnoremap <S-P> :GoImplements<cr>
" 按下Shift+R时，会执行Go代码的重命名操作
nnoremap <S-R> :GoRename<cr>
" 按下Shift+F时，会执行填充Go结构体的操作
nnoremap <S-F> :GoFillStruct<cr>
" 按下Shift+C时，会列出当前光标下标识符的调用者
nnoremap <S-C> :GoCallers<cr>
" 按下Shift+H时，会启用/禁用相同标识符的突出显示
nnoremap <S-H> :GoSameIdsToggle<cr>

let g:fzf_preview_window = ['hidden,right,50%,<70(up,40%)', 'ctrl-/']
" [Buffers] Jump to the existing window if possible
let g:fzf_buffers_jump = 1
" [[B]Commits] Customize the options used by 'git log':
let g:fzf_commits_log_options = '--graph --color=always --format="%C(auto)%h%d %s %C(black)%C(bold)%cr"'
" [Tags] Command to generate tags file
let g:fzf_tags_command = 'ctags -R'
```