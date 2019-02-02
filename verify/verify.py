#!/usr/bin/env python

import sys, os
import re
import calendar
import subprocess
import codecs
import argparse
import urllib.request, urllib.error
import shutil
import json
import time
import logging

TIMEOUT = 30

# Thanks for darkk
SIGNING_RE = re.compile(br'object: signingTime \(1\.2\.840\.113549\.1\.9\.5\)\s+(?:value\.)?set:\s+UTCTIME:(?P<mon>Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)\s+(?P<day>\d+) (?P<hour>\d\d):(?P<min>\d\d):(?P<sec>\d\d) (?P<year>\d{4}) GMT\s', re.DOTALL)
MONMAP = {b'Jan': 1, b'Feb': 2, b'Mar': 3, b'Apr': 4, b'May': 5, b'Jun': 6,
          b'Jul': 7, b'Aug': 8, b'Sep': 9, b'Oct': 10, b'Nov': 11, b'Dec': 12}

def cms_signing_time(cms):
        m = SIGNING_RE.search(cms)
        if m is None:
                raise RuntimeError('Signature file without signingTime')
        d = m.groupdict()
        for k in d.keys():
                if k == 'mon':
                        d[k] = MONMAP[d[k]] # NB, it's 1-12, not 0-11
                else:
                        d[k] = int(d[k], 10)
        return calendar.timegm((d['year'], d['mon'], d['day'], d['hour'], d['min'], d['sec']))

def confirm(apiurl, task, token, act):
        url = apiurl + "/task/" + task + "/" + act
        request = urllib.request.Request(url)
        request.add_header('Authorization', "Bearer %s" % token)
        request.get_method = lambda: 'PATCH'
        try:
                response = urllib.request.urlopen(request, timeout=TIMEOUT)
        except urllib.error.HTTPError as e:
                response = e
        except urllib.error.URLError as e:
                response = e
        #body = response.read()
        code = response.getcode()
        return code

def getq(apiurl, token):
        url = apiurl + "/queue"
        request = urllib.request.Request(url)
        request.add_header('Authorization', "Bearer %s" % token)
        response = urllib.request.urlopen(request, timeout=TIMEOUT)
        encoding = response.info().get_content_charset('utf-8')
        body = response.read()
        code = response.getcode()
        if code == 200:
                data = json.loads(body.decode(encoding))
                return code, data["id"]
        else:
                return code, ""

def cleanup(path):
        if os.path.exists(path):
                shutil.rmtree(path)


def handle(args, task):

        logger.info("Try to verify %s", task)
        path = os.path.join(args.datadir, task[0], task[1], task)

        if not os.path.exists(path) or not os.path.isdir(path):
                logger.warning("%s is not exists or not directory!", path)
                try:
                        code = confirm(args.apiurl, task, args.token, "fail")
                        logger.info("Confirm fail: %s", code)
                        if code in (200, 409, 400):
                                cleanup(path)
                except:
                        logger.error("Oops: %s", sys.exc_info()[1])

                return

        files = []
        c = 0
        for filename in os.listdir(path):
                if c == 3:
                        break
                files.append(filename)
                c += 1

        if len(files) != 2:
                logger.warning("Too many files!")
                try:
                        code = confirm(args.apiurl, task, args.token, "fail")
                        logger.info("Confirm fail: %s", code)
                        if code in (200, 409, 400):
                                cleanup(path)
                except:
                        logger.error("Oops: %s", sys.exc_info()[1])
                return

        if files[0].endswith(".sig"):
                sigfilename = files[0]
                datafilename = files[1]
        elif files[1].endswith(".sig"):
                sigfilename = files[1]
                datafilename = files[0]
        else:
                logger.error("Unknown files!")
                try:
                        code = confirm(args.apiurl, task, args.token, "fail")
                        logger.warning("Confirm fail: %s", code)
                        if code in (200, 409, 400):
                                cleanup(path)
                except:
                        logger.error("Oops: %s", sys.exc_info()[1])
                return

        logger.info("Data: %s signature: %s", datafilename, sigfilename)

        datafile = os.path.join(path, datafilename)
        sigfile = os.path.join(path, sigfilename)

        try:
                cms = subprocess.check_output(['openssl', 'pkcs7', '-inform', 'DER', '-in', sigfile, '-noout', '-print'])
                signing_ts = cms_signing_time(cms)

                oargs = ['openssl', 'smime', '-verify', '-noverify', '-engine', 'gost', '-CApath', args.capath, '-attime', str(signing_ts),
                        '-in', sigfile, '-inform', 'DER', '-content', datafile, '-out', '/dev/null']
                verify = subprocess.Popen(oargs , stderr=subprocess.PIPE)
                stderr = verify.stderr.read()
                if verify.wait() != 0 or b'Verification successful\n' not in stderr:
                        # `stderr` double check is needed because...
                        ### $ openssl smime -verify -engine gost -CApath /nonexistent -in dump.xml.sig -inform DER && echo OKAY
                        ### engine "gost" set.
                        ### smime: Not a directory: /nonexistent
                        ### smime: Use -help for summary.
                        ### OKAY <--- ^!(*&^%@(^%@#&$%!!!
                        # I hope, it has no messages like "Not Quite Verification successful\n"...
                        raise RuntimeError(' '.join(oargs), signing_ts, stderr)
                try:
                        with open(os.path.join(path,"confirm"),"w+") as f:
                                f.write("")
                        code = confirm(args.apiurl, task, args.token, "ok")
                        logger.info("Confirm ok: %s", code)
                        #if code in (200, 409, 400):
                        #        cleanup(path)
                except:
                        logger.error("Oops: %s", sys.exc_info()[1])
        except:
                logger.warning("Verify fail: %s", sys.exc_info()[1])
                try:
                        code = confirm(args.apiurl, task, args.token, "fail")
                        logger.info("Confirm fail: %s", code)
                        if code in (200, 409, 400):
                                cleanup(path)
                except:
                        logger.error("Oops: %s", sys.exc_info()[1])

if __name__ == "__main__":

        parser = argparse.ArgumentParser(description='Verify GOST documents')
        parser.add_argument('-d', dest="datadir", required=True, help='data directory')
        parser.add_argument('-c', dest="capath", required=True, help='GOST CA path')
        parser.add_argument('-u', dest="apiurl", required=True, help='API URL')
        parser.add_argument('-t', dest="token", required=True, help='Token')
        args = parser.parse_args()

        logger = logging.getLogger('main')
        logger.setLevel(logging.DEBUG)
        ch = logging.StreamHandler()
        ch.setLevel(logging.DEBUG)
        formatter = logging.Formatter('%(asctime)s - %(name)s - %(levelname)s - %(message)s')
        ch.setFormatter(formatter)
        logger.addHandler(ch)

        while True:
                dt = 0
                try:
                        code, task = getq(args.apiurl, args.token)
                        if code == 200:
                                logger.info("Found code: %s Task: %s", code, task)
                                handle(args, task)
                        if code == 204 and dt < 3:
                                dt += 1
                        else:
                                dt = 0
                except:
                        logger.error("Something wrong: %s", sys.exc_info()[1])
                        dt += 1
                if dt >=3:
                        dt = 3
                time.sleep(dt)

