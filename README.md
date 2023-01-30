# image/gif

A modified [image/gif](https://pkg.go.dev/image/gif)
package from the go standard library. The change is
to allow streaming gif to a file or to any io.writer.

## Purpose, motivation, rationale and other things wrong with my life

I was creating a shitty screen capture tool with
gif as an output. I left it running for testing
while watching my waifu do silly things, but
then it started freezing like crazy. Woah,
wut just happened I thought, then as I was
rattling my empty head for ideas, I just
remembered something really stupid: that
my other shitty library has a shitty memory
leak that was rather quite obvious in 20/20 sniper-esque
hindsight. I promptly fixed that, feeling
proud of myself that I managed to get 0 heap
allocations on the benchmark. Oh wait,
that isn't the root cause why my computer
froze. It's because the gif reached more
than 2000 images, and my shitty laptop
couldn't take it anymore.

Long story short, `image/gif` currently
doesn't have option to incrementally write to
a gif file without storing all the images in an
array first. So here I am, buried in yak
shavings, wondering what was doing in the
first place, the reason for my being,
why do we exist in the first place.
It's just a simple change, won't take
me an hour or two, says I the naive.
Hardly famous, unpopular
last words. It took me the rest of
the doing this. No worries though,
I have manage to implement it, and
get some test passing. What am I even
rambling about. I'm a little sore
from taking almost a day for such
a simple change. Why am I even rushing.
I have all the spare time in the world,
I have no places to be.
