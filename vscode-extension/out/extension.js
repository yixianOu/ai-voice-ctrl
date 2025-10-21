"use strict";
var __createBinding = (this && this.__createBinding) || (Object.create ? (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    var desc = Object.getOwnPropertyDescriptor(m, k);
    if (!desc || ("get" in desc ? !m.__esModule : desc.writable || desc.configurable)) {
      desc = { enumerable: true, get: function() { return m[k]; } };
    }
    Object.defineProperty(o, k2, desc);
}) : (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    o[k2] = m[k];
}));
var __setModuleDefault = (this && this.__setModuleDefault) || (Object.create ? (function(o, v) {
    Object.defineProperty(o, "default", { enumerable: true, value: v });
}) : function(o, v) {
    o["default"] = v;
});
var __importStar = (this && this.__importStar) || (function () {
    var ownKeys = function(o) {
        ownKeys = Object.getOwnPropertyNames || function (o) {
            var ar = [];
            for (var k in o) if (Object.prototype.hasOwnProperty.call(o, k)) ar[ar.length] = k;
            return ar;
        };
        return ownKeys(o);
    };
    return function (mod) {
        if (mod && mod.__esModule) return mod;
        var result = {};
        if (mod != null) for (var k = ownKeys(mod), i = 0; i < k.length; i++) if (k[i] !== "default") __createBinding(result, mod, k[i]);
        __setModuleDefault(result, mod);
        return result;
    };
})();
Object.defineProperty(exports, "__esModule", { value: true });
exports.activate = activate;
exports.deactivate = deactivate;
const vscode = __importStar(require("vscode"));
const http = __importStar(require("http"));
let server = null;
function activate(context) {
    console.log('AI Voice Control Bridge activated');
    const config = vscode.workspace.getConfiguration('aiVoiceCtrlBridge');
    const autoStart = config.get('autoStart', true);
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
function startServer(context) {
    if (server) {
        vscode.window.showWarningMessage('Bridge server already running');
        return;
    }
    const config = vscode.workspace.getConfiguration('aiVoiceCtrlBridge');
    const port = config.get('port', 9527);
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
            }
            catch (error) {
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
async function handleCommand(command) {
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
async function openFile(params) {
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
async function closeFile(params) {
    const uri = vscode.Uri.file(params.path);
    const editors = vscode.window.visibleTextEditors.filter(e => e.document.uri.toString() === uri.toString());
    for (const editor of editors) {
        await vscode.window.showTextDocument(editor.document);
        await vscode.commands.executeCommand('workbench.action.closeActiveEditor');
    }
    return { success: true, message: 'File closed', path: params.path };
}
async function refreshFile(params) {
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
async function closeWindow() {
    await vscode.commands.executeCommand('workbench.action.closeWindow');
    return { success: true, message: 'Window closed' };
}
async function executeCommand(params) {
    const result = await vscode.commands.executeCommand(params.command, ...(params.args || []));
    return { success: true, message: 'Command executed', result };
}
function deactivate() {
    stopServer();
}
//# sourceMappingURL=extension.js.map