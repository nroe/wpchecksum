wpchecksum
==========

检查指定目录下的 WordPress 程序是否有被修改过

    ver        需要对比的 WordPress 版本号
    dir        检查目录
    checksum   校验数据保存地址，如果不指定会下载保存到临时目录。如果指定，第一次将写入数据，以后将直接读取

### 样例：

检查 ./WordPress-3.9 目录下资源 和 3.9 版本的差异。并且创建文件 `3.9.checksum.csv` 保存 checksum 校验数据
`#./release/darwin_386/wpchecksum --ver=3.9 --dir=./WordPress-3.9/ --checksum=3.9.checksum.csv`
>        start check directory './WordPress-3.9' ...
>        result:
>            file:  1246
>            diff:  0
>            lost:  0

### 截图：

![screenshot-1](https://raw.githubusercontent.com/nroe/wpchecksum/master/screenshot/2014-05-12-11.01.56-AM.jpg)
