import { APP_BASE_HREF } from '@angular/common';
import { CommonEngine } from '@angular/ssr';
import express from 'express';
import { fileURLToPath } from 'node:url';
import { dirname, join, resolve } from 'node:path';
import bootstrap from './src/main.server';

// Base href - se puede sobreescribir con variable de entorno
// Valores: '/' para producción (llanteradeoccidente.com), '/llantera/' para staging (ikingarisolorzano.com/llantera)
const BASE_HREF = process.env['BASE_HREF'] || '/';

// The Express app is exported so that it can be used by serverless Functions.
export function app(): express.Express {
    const server = express();
    const serverDistFolder = dirname(fileURLToPath(import.meta.url));
    const browserDistFolder = resolve(serverDistFolder, '../browser');
    const indexHtml = join(serverDistFolder, 'index.server.html');

    const commonEngine = new CommonEngine();

    server.set('view engine', 'html');
    server.set('views', browserDistFolder);

    // Example Express Rest API endpoints
    // server.get('/api/**', (req, res) => { });
    // Serve static files from /browser
    server.get('**', express.static(browserDistFolder, {
        maxAge: '1y',
        index: 'index.html',
    }));

    // All regular routes use the Angular engine
    server.get('**', (req, res, next) => {
        const { protocol, originalUrl, headers } = req;

        // Reconstruir la URL completa incluyendo el BASE_HREF
        // Nginx con trailing slash elimina el prefijo, así que lo agregamos de vuelta
        const baseHrefPath = BASE_HREF.endsWith('/') ? BASE_HREF.slice(0, -1) : BASE_HREF;
        const fullUrl = baseHrefPath === '' || originalUrl.startsWith(baseHrefPath)
            ? `${protocol}://${headers.host}${originalUrl}`
            : `${protocol}://${headers.host}${baseHrefPath}${originalUrl}`;

        commonEngine
        .render({
            bootstrap,
            documentFilePath: indexHtml,
            url: fullUrl,
            publicPath: browserDistFolder,
            providers: [{ provide: APP_BASE_HREF, useValue: BASE_HREF }],
        })
        .then((html) => res.send(html))
        .catch((err) => next(err));
    });

    return server;
}

function run(): void {
    const port = process.env['PORT'] || 4000;

    // Start up the Node server
    const server = app();
    server.listen(port, () => {
        console.log(`Node Express server listening on http://localhost:${port}`);
    });
}

run();