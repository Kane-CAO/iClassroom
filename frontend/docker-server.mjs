import { createReadStream } from 'node:fs'
import { stat } from 'node:fs/promises'
import { createServer } from 'node:http'
import { extname, join, normalize } from 'node:path'

const root = join(process.cwd(), 'dist')
const host = '0.0.0.0'
const port = Number(process.env.PORT || 4173)

const contentTypes = {
  '.css': 'text/css; charset=utf-8',
  '.html': 'text/html; charset=utf-8',
  '.ico': 'image/x-icon',
  '.js': 'text/javascript; charset=utf-8',
  '.json': 'application/json; charset=utf-8',
  '.map': 'application/json; charset=utf-8',
  '.png': 'image/png',
  '.svg': 'image/svg+xml',
  '.txt': 'text/plain; charset=utf-8',
  '.webp': 'image/webp',
}

createServer(async (req, res) => {
  const url = new URL(req.url || '/', `http://${req.headers.host || 'localhost'}`)
  const requested = normalize(decodeURIComponent(url.pathname)).replace(/^(\.\.[/\\])+/, '')
  const filePath = await resolveFile(requested)
  const type = contentTypes[extname(filePath)] || 'application/octet-stream'

  res.setHeader('Content-Type', type)
  createReadStream(filePath)
    .on('error', () => {
      res.statusCode = 500
      res.end('Internal Server Error')
    })
    .pipe(res)
}).listen(port, host, () => {
  console.log(`iClassroom frontend listening on http://${host}:${port}`)
})

async function resolveFile(pathname) {
  const candidate = join(root, pathname === '/' ? 'index.html' : pathname)
  if (candidate.startsWith(root)) {
    try {
      const info = await stat(candidate)
      if (info.isFile()) {
        return candidate
      }
    } catch {
      // Fall through to the SPA entry.
    }
  }
  return join(root, 'index.html')
}
