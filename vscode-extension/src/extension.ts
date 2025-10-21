import * as vscode from 'vscode';
import * as http from 'http';

let server: http.Server | null = null;

export function activate(context: vscode.ExtensionContext) {
    console.log('AI Voice Control Bridge activated');

    const config = vscode.workspace.getConfiguration('aiVoiceCtrlBridge');
    const autoStart = config.get<boolean>('autoStart', true);

    const startCommand = vscode.commands.registerCommand('ai-voice-ctrl-bridge.startServer', () => {
        startServer(context);
    });

    const stopCommand = vscode.commands.registerCommand('ai-voice-ctrl-bridge.stopServer', () => {
        stopServer();
    });

    context.subscriptions.push(startCommand, stopCommand);

    if (autoStart) {
        startServer(context);
    }
}

function startServer(context: vscode.ExtensionContext) {
    if (server) {
        vscode.window.showWarningMessage('Bridge server already running');
        return;
    }

    const config = vscode.workspace.getConfiguration('aiVoiceCtrlBridge');
    const port = config.get<number>('port', 9527);

    server = http.createServer(async (req, res) => {
        res.setHeader('Content-Type', 'application/json');
        res.setHeader('Access-Control-Allow-Origin', '*');
        res.setHeader('Access-Control-Allow-Methods', 'POST, OPTIONS');
        res.setHeader('Access-Control-Allow-Headers', 'Content-Type');

        if (req.method === 'OPTIONS') {
            res.writeHead(200);
            res.end();
            return;
        }

        if (req.method !== 'POST') {
            res.writeHead(405);
            res.end(JSON.stringify({ success: false, message: 'Method not allowed' }));
            return;
        }

        let body = '';
        req.on('data', chunk => body += chunk.toString());
        req.on('end', async () => {
            try {
                const command = JSON.parse(body);
                const result = await handleCommand(command);
                res.writeHead(200);
                res.end(JSON.stringify(result));
            } catch (error) {
                res.writeHead(400);
                res.end(JSON.stringify({ 
                    success: false, 
                    message: error instanceof Error ? error.message : 'Unknown error' 
                }));
            }
        });
    });

    server.listen(port, () => {
        vscode.window.showInformationMessage(`Bridge server started on port ${port}`);
        console.log(`Bridge server listening on port ${port}`);
    });
}

function stopServer() {
    if (!server) {
        vscode.window.showWarningMessage('Bridge server not running');
        return;
    }

    server.close(() => {
        vscode.window.showInformationMessage('Bridge server stopped');
        console.log('Bridge server stopped');
    });
    server = null;
}

async function handleCommand(command: any): Promise<any> {
    const { action, params } = command;

    switch (action) {
        case 'open_file':
            return await openFile(params);
        case 'close_file':
            return await closeFile(params);
        case 'refresh_file':
            return await refreshFile(params);
        case 'close_window':
            return await closeWindow();
        case 'execute_command':
            return await executeCommand(params);
        default:
            throw new Error(`Unknown action: ${action}`);
    }
}

async function openFile(params: { path: string; line?: number }): Promise<any> {
    const uri = vscode.Uri.file(params.path);
    const doc = await vscode.workspace.openTextDocument(uri);
    const editor = await vscode.window.showTextDocument(doc);

    if (params.line !== undefined && params.line > 0) {
        const position = new vscode.Position(params.line - 1, 0);
        editor.selection = new vscode.Selection(position, position);
        editor.revealRange(new vscode.Range(position, position), vscode.TextEditorRevealType.InCenter);
    }

    return { success: true, message: 'File opened', path: params.path };
}

async function closeFile(params: { path: string }): Promise<any> {
    const uri = vscode.Uri.file(params.path);
    const editors = vscode.window.visibleTextEditors.filter(e => e.document.uri.toString() === uri.toString());

    for (const editor of editors) {
        await vscode.window.showTextDocument(editor.document);
        await vscode.commands.executeCommand('workbench.action.closeActiveEditor');
    }

    return { success: true, message: 'File closed', path: params.path };
}

async function refreshFile(params: { path: string }): Promise<any> {
    const uri = vscode.Uri.file(params.path);
    const doc = vscode.workspace.textDocuments.find(d => d.uri.toString() === uri.toString());

    if (doc && doc.isDirty) {
        await doc.save();
    }

    // Close and reopen to force refresh
    await closeFile(params);
    await new Promise(resolve => setTimeout(resolve, 100));
    await openFile(params);

    return { success: true, message: 'File refreshed', path: params.path };
}

async function closeWindow(): Promise<any> {
    await vscode.commands.executeCommand('workbench.action.closeWindow');
    return { success: true, message: 'Window closed' };
}

async function executeCommand(params: { command: string; args?: any[] }): Promise<any> {
    const result = await vscode.commands.executeCommand(params.command, ...(params.args || []));
    return { success: true, message: 'Command executed', result };
}

export function deactivate() {
    stopServer();
}
