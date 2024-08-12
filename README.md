# incant

Incant is a static site generator (SSG) similar to Hugo or Zola. However it aims for a less-opinionated, more customizable approach.

Fundamentally, an SSG combines together...
- some structure content that describes the site
- with some templates that describe how the structure content corresponds to the output HTML
- according to some content<->template relationship

Incant tries to make these three components very explicit, and otherwise it tries to provide complete flexibility. For example:
- Zola requires that site items be separated into their own files
- The item file format supports a specific collection of supported keywords, each with specific purposes
- If you want an item to have multiple markdown content blurbs, this isn't possible. Or at least it's not the standard workflow.

By comparison, besides the top-level site configuration file, `incant` allows for any content structure that makes sense for you. The only fully-structured file is the top-level configuration file:
```
{
    ContentRoot: ./content
    SiteDataFile: site_data.hjson
    MappingFile: mapping.hjson
    TemplatesRoot: ./templates
    StaticRoot: ./static
    OutputRoot: ./output
}
```

In this example:
- the `site_data.hjson` file describes the site content.
- the `./templates` directory contains a library of templates.
- the `mapping.hjson` file describes which templates should be applied to which parts of the content.

## Interesting tidbits
- `jq` syntax is used in the mapping to select subsets of the total site content.
- We started by supporting .toml-based configuration, but we ran into limitations. Then we tried .yaml. JSON5. And finally HJSON. The good news is: You can use any of these that you like. The file loader can load any of these formats, and deserializes them into an in-memory, agnostic format. If there's another format you're interested in, let us know!

## Usage
```
go run . --config=example/config.hjson
```

## Disclaimer
`incant` isn't especially full-featured yet. There are some yucky bits even in common functionality, like creating links between different parts of the site. We're working on it!

Also, `incant` has the deliberate disadvantage that its customizability and flexibility will make it difficult or impossible for style kits to be shared between sites, unless they also enforce a common site content structure. 

## In conclusion...
If anything about `incant` sounds interesting, let us know!
