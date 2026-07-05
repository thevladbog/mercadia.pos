// Minimal Tauri v2 desktop shell for the pos-terminal app.
//
// No custom commands are registered here (see plan 013): devices talk ONLY to the
// Hardware Agent over HTTP, the shell must not become a second device layer.

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
