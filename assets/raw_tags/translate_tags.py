#!/usr/bin/env python3.4

import sys

import ruamel.yaml


class HexInt(int):
    pass

def representer(dumper, data):
    return \
        ruamel.yaml.ScalarNode(
            'tag:yaml.org,2002:int',
            '0x{:04x}'.format(data))

ruamel.yaml.add_representer(HexInt, representer)

data = {
    'item1': {
        'string_value': 'some_string',
        'hex_value': HexInt(641),
    }
}

ruamel.yaml.dump(data, sys.stdout, default_flow_style=False)

sys.exit(0)

#!/usr/bin/env python2.7

"""
Parses the table-data from view-source:http://www.exiv2.org/tags.html
"""

import sys
import collections

import xml.etree.ElementTree as ET

import ruamel.yaml


# Prepare YAML to write hex expressions (otherwise the hex will be a string and
# quotes or a decimal and a base-10 number).

class HexInt(int):
    pass

def representer(dumper, data):
    return \
        ruamel.yaml.ScalarNode(
            'tag:yaml.org,2002:int',
            '0x{:04x}'.format(data))

ruamel.yaml.add_representer(HexInt, representer)

data = {
    'item1': {
        'string_value': 'some_string',
        'hex_value': HexInt(641),
    }
}

ruamel.yaml.dump(data, sys.stdout, default_flow_style=False)



sys.exit(0)

def _write(tags):
    writeable = {}

    for tag in tags:
        pivot = tag['fq_key'].rindex('.')

        item = {
            'id': HexInt(tag['id_dec']),
            'name': tag['fq_key'][pivot + 1:],
        }

        try:
            writeable[tag['ifd']].append(item)
        except KeyError:
            writeable[tag['ifd']] = [item]

    with open('tags.yaml', 'w') as f:
        # Otherwise, the next dictionaries will look like Python dictionaries,
        # whatever sense that makes.
        ruamel.yaml.dump(writeable, f, default_flow_style=False)

def _main():
    tree = ET.parse('tags.html')
    root = tree.getroot()

    labels = [
        'id_hex',
        'id_dec',
        'ifd',
        'fq_key',
        'type',
        'description',
    ]

    tags = []
    for node in root.iter('tr'):
        values = [child.text.strip() for child in node.iter('td')]

        # Skips the header row.
        if not values:
            continue

        assert \
            len(values) == len(labels), \
            "Row fields count not the same as labels: {}".format(values)

        tags.append(dict(zip(labels, values)))

    _write(tags)

if __name__ == '__main__':
    _main()
