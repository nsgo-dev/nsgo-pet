# Shimeji

From [Shimeji for Linux](https://github.com/estenv/linux-shimeji) at `b90492d326cb24b9ceba914481a78a5e2470c0dc` (`img/shime*.png`), a port of the original [Shimeji](http://www.group-finity.com/Shimeji) by Yuki Yamada. Its README says: "This project inherits the ZLIB/LIBPNG license of the original program."

**This is an altered version.** The images are cropped, scaled and encoded as video clips by `tools/framepet` from `tools/framepet/pets/shimeji.json`. The clips follow the poses and lengths of `conf/Actions.xml`: 落下する (drop-in), ジャンプ (jump), 立つ (stand), 座る (sit), 座って見上げる (look-up), 座って首が回る (head-spin), 寝そべる (sprawl), ずりずり played in place (creep), 転ぶ without its last pose (trip), 跳ねる (bounce), 分裂1 (split), and ドラッグされる (pinched). They play in nsgo-pet, not in Shimeji, and are not the original software.

Its [LICENSE](https://github.com/estenv/linux-shimeji/blob/b90492d326cb24b9ceba914481a78a5e2470c0dc/LICENSE), unaltered:

```text
Copyright (c) 2009-2011 Yuki Yamada

This software is provided 'as-is', without any express or implied warranty. In no event will the authors be held liable for any damages arising from the use of this software.

Permission is granted to anyone to use this software for any purpose, including commercial applications, and to alter it and redistribute it freely, subject to the following restrictions:

1. The origin of this software must not be misrepresented; you must not claim that you wrote the original software. If you use this software in a product, an acknowledgment in the product documentation would be appreciated but is not required.

2. Altered source versions must be plainly marked as such, and must not be misrepresented as being the original software.

3. This notice may not be removed or altered from any source distribution.

-----------------------------------------



```
