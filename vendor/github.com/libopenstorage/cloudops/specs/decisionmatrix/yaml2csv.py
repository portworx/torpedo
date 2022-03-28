#!/usr/bin/env python3

import sys
import os
import yaml


def yaml2csv(fname):
    ''' Converts drive-YAML syntax to CSV '''
    f = open(fname)
    a = yaml.load(f)
    f.close()

    printOrder = [
        'drive_type',		    # 'gp2',
        'region',			    # '*',
        'instance_type',		# '*',
        'priority',			    # 0,
        'thin_provisioning',  # False
        'instance_min_drives',  # 1,
        'instance_max_drives',  # 8,
        'min_iops',			    # 100,
        'max_iops',			    # 100,
        'min_size',			    # 0,
        'max_size',			    # 33,
    ]

    print(','.join(printOrder))
    for r in a['rows']:
        print(','.join(str(r[_]) for _ in printOrder))


def csv2yaml(fname):
    ''' Converts drive data from CSV to YAML '''
    f = open(fname)
    LABELS = f.readline()[:-1].split(',')

    print('rows:')
    for line in f:
        prefix = '-'
        parts = line[:-1].split(',')
        for i in range(len(parts)):
            p = parts[i].lower()
            if p == '*':
                p = "'{0}'".format(p)
            print('{0} {1}: {2}'.format(prefix, LABELS[i], p))
            prefix = ' '
    f.close()


if __name__ == '__main__':
    for fname in sys.argv[1:]:
        _, ext = os.path.splitext(fname)
        if ext == '.yaml':
            yaml2csv(fname)
        elif ext == '.csv':
            csv2yaml(fname)
        else:
            raise ValueError('unhandled extension:'+ext)
