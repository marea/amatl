# amatl

**amatl** is a minimalistic static site generator buillt with Go.

## Installation

```bash
go install github.com/marea/amatl
```

## Usage

```bash
amatl
```

On first use it will generate the following directories:

```
assets/
dist/
inc/
```

### Assets

This folder contains all your static files.

They will be copied over to the dist folder with the same folder structure.

### Dist

This folder will contain your compiled site.

### Inc

This folder contains all your pages, they must have the `.html` extension.

They will be copied over to the dist folder with the same folder structure.

Pages will use their file name without extension as their title, if you want to
change it, you can add a `# title:My Title` tag on the first line of the file.

## Contributing

Pull requests are welcome. For major changes, please open an issue first to
discuss what you would like to change.

Please make sure to update tests as appropriate.

## License

[Unlicense](https://choosealicense.com/licenses/unlicense/)
