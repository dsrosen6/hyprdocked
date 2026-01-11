# hyprdocked

`hyprdocked` is a laptop display helper for Hyprland that updates the display state based on a few factors.

## How it Works

`hyprdocked` listens for the following events:

- Displays being added or removed (via [Hyprland's IPC](https://wiki.hypr.land/IPC/))
- Laptop lid events (when it is opened or closed)
- Laptop power events (changes in AC or battery state)
- `hyprdocked suspend` or `hyprdocked wake` events (keep reading for details on this)

Any time one of the above events are received, `hyprdocked` applies settings based on the following statuses if changes are needed.

The laptop display is *enabled* if the following statuses are detected:

- Docked with lid opened
- Laptop only (any lid status)

The laptop display is *disabled* if the device is detected as docked with lid closed.

### Special Case: `hyprdocked suspend`

If the command `hyprdocked suspend` is called, the laptop display is enabled (regardless of the above statuses) and is kept that way until `hyprdocked wake` is called to release it.

The reason for this is because otherwise, when Hyprland is suspended, it is in whatever state it was last in until it wakes back up.

Why? If your laptop is suspended (and presumably locked) while docked (manually or via an idle agent) and then unplugged from the dock, then you're opening up your laptop but the laptop display is still disabled. And since you now have zero displays enabled, you get the dreaded "oopsie daisy" screen on Hyprland.

This needs some manual wiring, because `hyprdocked` does not assume which idle utility you use. See the `Idle Daemon` section in configuration.

## Installation

### From Source

Just run `go install github.com/dsrosen6/hyprdocked@latest` and then move to configuration.

### NixOS (Home Manager)

In your `flake.nix`:

```nix
# The beginning of your flake...

inputs = {
    hyprdocked.url = "github:dsrosen6/hyprdocked";
    # your other inputs...
};

outputs = {
    hyprdocked,
    # your other outputs...
    ...
};

# the rest of your flake...
```

In your `home.nix`, or wherever you put it:  

```nix
{
    inputs,
    # other variables...
    ...
}:
{
    imports = [
        inputs.hyprdocked.homeManagerModules.default
        # your other imports...
    ];

    # Your stuff...

    services.hyprdocked.enable = true;
}
```

You can skip "auto-run" in configuration.

## Configuration

`hyprdocked` requires very minimal configuration on top of your existing Hyprland config:

### Auto-Run

If you're running Hyprland with UWSM:

1. Download the file `hyprdocked.service` from this repo
2. Add the file to `~/.config/systemd/user/`
3. Run `systemctl --user daemon-reload`
4. Run `systemctl --user enable hyprdocked.service --now`

Otherwise, you can add to your Hyprland config:
`exec-once = hyprdocked`

### Identify Laptop Display

Run `hyprctl monitors` and find your laptop display. If it is anything like eDP-1, eDP1, you can just skip to the next section. To know if this applies to yours, just turn the name into full lowercase and remove the dash. If it is `edp1`, you're good.

If it's not, set an environment variable of `LAPTOP_DISPLAY_NAME` with the value being whatever you found.

*If you need to do this, raise an issue. I'm happy to add  it to the common auto-detected display names.*

### Hyprland Monitors

Make sure your monitors in your Hyprland config are all set as enabled. This ensures that `hyprdocked` can properly capture the right settings on startup. ***At an absolute minimum, put your laptop display settings in your config.***

For example:

```conf
monitor = DP-1,3440x1440@174.96,0x0,1.0 # external
monitor = eDP-1,1920x1200,3440x0,1.25 # laptop, required
```

### Idle Daemon

Assuming you're using `hypridle`, you need to do the following.

1. Add `hyprdocked suspend` into your `before_sleep_cmd`.
2. Add `hyprdocked wake` into your `after_sleep_cmd`.

Here is an example of how my `hypridle` config looks:

```conf
general {
    before_sleep_cmd=sh -c 'loginctl lock-session && hyprdocked suspend'
    after_sleep_cmd=hyprdocked wake
    lock_cmd=pidof hyprlock || hyprlock
}

listener {
    # Suspend after 30 seconds if hyprlock is active. Good for manual
    # locks and accidental wakes.
    timeout = 30
    on-timeout = sh -c 'pidof hyprlock >/dev/null && systemctl suspend'
}

listener {
    on-timeout=loginctl lock-session
    timeout=600
}

listener {
    on-timeout=systemctl suspend
    timeout=630
}
```

Translate to your idle agent if you use a different one.
