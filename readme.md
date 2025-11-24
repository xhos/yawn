# yawn - a very boring greeter 💤

Minimal, sleek [greetd](https://git.sr.ht/~kennylevinsen/greetd) based greeter written in Go with [Bubble Tea](https://github.com/charmbracelet/bubbletea).

It's simple, it's dumb, it just works.

## Configuration

Configuration is done via command-line flags:

| Flag       | Default | Description                         |
|------------|---------|-------------------------------------|
| `-cmd`     | none    | Command to run for user session     |
| `-preauth` | `false` | Start the auth loop immediately     |
| `-user`    | none    | Force a specific username           |
| `-width`   | `8`     | Width of the input fields           |

## Why another greeter?

> **y**et **a**nother t**w**ui greet**n**er

Initially built for a single-user NixOS + Hyprland setup with a few specific requirements:

1. **Force username** - No need to type it every boot (since I don't want any state like last logged in user)
2. **Instant fingerprint auth** - Start the PAM loop immediately with `-preauth`, no extra keypresses
3. **Simple** - Just a TUI and a few flags

If you have a laptop with a fingerprint reader and want seamless auth without the "type username -> press enter -> now you can use fingerprint" dance, this is for you.

## Known limitations

Multi-monitor scaling is broken - that's just how TTYs work. For example, I have a 4k display and a 1080p display. The greeter renders correctly on the 1080p, but it's tiny on the 4k one. This affects all TUI greeters. The only workaround is launching a graphical environment (Cage + terminal), but that defeats the simplicity goal. [sysc-greet](https://github.com/Nomadcxx/sysc-greet) takes that approach if it bothers you too much.

## Usage

### NixOS

Add to your flake inputs:

```nix
{
  inputs.yawn.url = "github:xhos/yawn";
}
```

Then configure greetd:

```nix
services.greetd = {
  enable = true;
  settings.default_session.command = "${inputs.yawn.packages.${pkgs.system}.default}/bin/yawn -cmd Hyprland";
};
```

### Other distros

`/etc/greetd/config.toml`:

```toml
[terminal]
vt = 1

[default_session]
command = "yawn -user xhos -cmd Hyprland -preauth"
user = "greeter"
```

Refer to [greetd's docs](https://man.sr.ht/~kennylevinsen/greetd/) for more details.

## Development

### VM

Launch a VM with greetd and yawn preconfigured:

```bash
nix run .#test-vm
```

To adjust args, edit the `test-vm` app in `flake.nix`.

### Local

For quick UI iteration without spawning a full VM, use the stub greetd server:

```bash
nix develop # if you don't use devenv
stub -- -cmd Hyprland -user test
```

The stub uses `test/test` as credentials.

## Similar projects

Other greeters worth checking out, these are the ones I tried before making yawn:

- [tuigreet](https://github.com/apognu/tuigreet)
- [regreet](https://github.com/rharish101/ReGreet)
- [sysc-greet](https://github.com/Nomadcxx/sysc-greet)
- [ly](https://github.com/fairyglade/ly)
- [lemurs](https://github.com/coastalwhite/lemurs)
- [sddm](https://github.com/sddm/sddm)
- [lightdm](https://github.com/canonical/lightdm)

## Contributing

PRs and issues are welcome. Packaging for other distros is especially appreciated since I only use NixOS.
