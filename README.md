# ReeeMiks

ReeeMiks is a fork of deej and ReeeMiks is an open source software volume mixer, with easy to build open source hardware, this fork is mainly focused on Linux but Windows should continue to function.

[![Discord](https://img.shields.io/discord/702940502038937667?logo=discord)](https://discord.gg/X7gScNud)

I'll put more into this later, for now reference deej if you have questions, if you'd like more tailored support for ReeeMiks then please [join the discord](https://discord.gg/X7gScNud).


## Why fork?

Because of the actions of the previous developer, I've decided that people shouldn't have to wait over 3 years for any kind of new features or potential bug fixes, I'm open to accepting PRs from anyone.

I'd rather take on new features with the community who want a virtual volume mixer than to wait for something that might not be coming.

### What does that mean?

I'm open to collaborate with people on getting new features and bug fixes into a more central repository rather than leaving it to the general user to try and merge every fork to get the features they want.

This does mean that I'm open to giving people who have existing forks with sizable new features (and that I can trust) access to approve and merge PRs as well. If you're one of those people, please join the discord and we can talk about getting you access if you meet the criteria.

ReeeMiks is also something which could have a much larger developer community around it, which I think would be great for everyone.


## Ok, but what are the new features so far?

If there is missing information or the feature could be expanded on or explained better, please open an issue and we'll talk about how to make the below feature explanations better.

1. Better support for pipewire by allowing devices to be controlled directly instead of only the default sound device.

This is achieved through pipewire's pulseaudio interface but allows ReeeMiks to reference any arbitrary (virtual or real) pipewire soundcards.
The naming convention is a little difficult to understand right now but if you use the development builds, it will list the devices and applications that ReeeMiks can see. This allows you to copy and paste those names into your configuration, so you can be sure that you can control the device/application you want to.

2. HID and serial support.

I've merged support for using HID (via qmk) while retaining support for serial (if you'd rather use the provided ReeeMiks arduino code, as it has new features too)

3. Button/key support (serial only, not available via HID mode).

This allows the use of key switches or buttons to trigger certain actions from your deej device via ReeeMiks' host application.

4. Automatic serial reconnection/retries (this does not work for HID mode).

If you unplug your deej hardware, ReeeMiks will try to reconnect automatically for you, there may be bugs with this if the serial port changes from what is provided in the config. If you find issues with this feature, open an issue and we can fix it together.

5. A working systray.

Deej's Go binary tries to use a no longer functioning systray library, which has since been replaced by a maintained one. I've made the changes to the software to allow it to build and function with the new library.

6. Reduced CPU utilisation (serial only, HID mode untested).

ReeeMiks waits for more data to be sent from the serial port instead of the arduino constantly sending the current value, which means that ReeeMiks only receives data when the deej hardware has actually new information to tell it. More battery life and less CPU contention as this removes the hot-loop that deej has.

7. Better noise reduction on the hardware side (serial only).

As part of the reduced CPU utilisation, ReeeMiks reduces the amount of updates sent over the serial connection by making sure that the value from the pins on the microcontroller are "stable" and not produced by noise.

ReeeMiks' arduino code does this via a history based debounce algorithm and in theory should make noisy hardware less noisy by pretty much asking the pin "are you really sure about that value?" a few times in a row, if the value is consistent enough, ReeeMiks will send that new value over the serial port.

The host software still retains the option to change the noise level via the existing host side debounce for easy reconfiguration without having to reflash the arduino but it means we don't need to inform the host machine as much and this is how we reduce the CPU that ReeeMiks uses.


## This sounds good but how do I get started?
ReeeMiks still works with existing deej hardware that is flashed with the arduino code from deej, but you'll be missing out on a few of the features above if you don't reflash with ReeeMiks' arduino code.

Other than the arduino based features, ReeeMiks has 100% of the same hardware support and software/hardware build steps as deej, so you're free to reference deej's software/hardware build guides as ReeeMiks doesn't care about what kind of hardware you have, however please try to rule out hardware issues on the arduino as this can easily cause issues.

## This seems more complicated than deej

I can understand that ReeeMiks will be harder for people to understand in comparison to deej, thats ok and I'm open to fixing that, if you need help building a device [join the discord](https://discord.gg/X7gScNud) and ask for help.

While this project might just be me right now, I'd like to get more people on board with the project and interacting with each other so we can help each other out.

If someone would like to write FAQs or help with documentation or submit PRs or suggest new features, then [join the discord](https://discord.gg/X7gScNud), open PRs and GitHub issues, together we can work it out.

