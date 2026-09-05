(function () {
  const translations = {
    pt: {
      tagline: 'Acesse seus arquivos',
      tabLogin: 'Login',
      tabRegister: 'Cadastro',
      labelCode: 'Código',
      labelPassword: 'Senha',
      labelName: 'Nome',
      placeholderCode: 'ex: abc123',
      placeholderName: 'Seu nome completo',
      btnLogin: 'Entrar',
      btnLoginLoading: 'Entrando...',
      btnRegister: 'Criar Conta',
      btnRegisterLoading: 'Criando...',
      footerNoAccount: 'Não tem uma conta?',
      footerSignUp: 'Cadastre-se',
      footerHasAccount: 'Já tem uma conta?',
      footerSignIn: 'Faça login',
      msgLoginSuccess: '✅ Login realizado com sucesso!',
      msgLoginErrorDefault: 'Código ou senha incorretos',
      msgRegisterSuccess: '✅ Conta criada com sucesso! Faça login.',
      msgRegisterErrorDefault: 'Erro ao criar conta',
      msgConnectionError: '❌ Erro ao conectar: ',
      uploadPageTitle: 'Upload - Wireless Drive',
      mediaPageTitle: 'Mídia - Wireless Drive',
      uploadTitle: 'Fazer Upload',
      uploadSubtitle: 'Arraste arquivos aqui ou selecione para fazer upload (múltiplos arquivos suportados)',
      dropText: 'Arraste seus arquivos aqui',
      dropSubtext: 'ou',
      selectFilesButton: 'Selecione arquivos',
      filesSelected: 'Arquivos selecionados',
      removeFile: 'Remover',
      uploadButton: 'Enviar Uploads 🚀',
      tipsTitle: 'Dicas',
      selectAtLeastOneFile: '❌ Selecione pelo menos um arquivo',
      needLoginToUpload: 'Você precisa entrar para enviar arquivos. Redirecionando...',
      uploading: 'Enviando...',
      uploadSupportedFormats: 'Formatos suportados: JPG, PNG, GIF, MP4, WebM, PDF, DOC, DOCX, PPT, XLS, TXT, ZIP',
      uploadImageMaxSize: 'Tamanho máximo recomendado para imagens: 50MB',
      uploadAutoTitle: 'Os títulos serão preenchidos automaticamente com o nome do arquivo e podem ser editados depois',
      uploadMultipleFiles: 'Você pode fazer upload de múltiplos arquivos simultaneamente',
      generatingThumbnail: 'Gerando Thumbnail 🖼️',
      uploadedSuccess: '✅ Enviado com sucesso',
      uploadSuccessAll: '✅ Todos os {count} arquivo(s) foram enviados com sucesso!',
      uploadPartialSuccess: '⚠️ {uploaded} arquivo(s) enviado(s) com sucesso, mas {failed} falharam.',
      uploadFailedAll: '❌ Falha ao enviar {count} arquivo(s).',
      sessionExpired: '❌ Sessão expirada',
      errorUpload: '❌ Erro: {error}',
      errorConnection: '❌ Erro: falha de conexão',
      unknownError: 'Erro desconhecido',

      // Dashboard
      navUpload: '📤 Upload',
      navLogout: '🚪 Sair',
      searchPlaceholder: 'Pesquisar mídias...',
      searchClearAria: 'Limpar pesquisa',
      toastCloseAria: 'Fechar',
      btnGenerateThumbnails: '🖼️ Gerar Thumbnails',
      btnGeneratingThumbnails: 'Gerando...',
      viewToggleToCompact: 'Ver como Grade Compacta',
      viewToggleToGrid: 'Ver como Grid',
      sortByDate: 'Data',
      sortAsc: 'A→Z',
      sortDesc: 'Z→A',
      noResultsFor: 'Nenhuma mídia encontrada para',
      noMediaYet: 'Sem mídias ainda. Faça um',
      uploadCta: 'upload',
      loadMediaError: 'Erro ao carregar mídias.',
      lightboxLoading: 'Carregando mídia...',
      lightboxLoadingSub: 'Aguarde enquanto a mídia é transferida',
      retryBtn: '↻ Tentar Novamente',
      closeBtn: 'Fechar',
      errorTitle: 'Erro ao Carregar',
      errorVideo: 'Não foi possível carregar o vídeo. Tente novamente.',
      errorMedia: 'Não foi possível carregar a mídia. Verifique sua conexão de internet e tente novamente.',
      allThumbnailsExist: 'Todas as mídias já possuem thumbnails!',
      errorGeneratingThumbnailFor: 'Erro ao gerar thumbnail para',
      errorGeneratingThumbnails: 'Impossível gerar thumbnails, servidor indisponível ou erro inesperado',
      authForbidden: 'Você não tem permissão para acessar isso. Redirecionando para o login...',
      authExpired: 'Sua sessão expirou. Redirecionando para o login...',

      // Media page
      navShuffle: 'Aleatório',
      backLink: '← Voltar',
      deleteLink: '🗑️ Excluir',
      loadingTitle: 'Carregando...',
      loadingDate: '📅 Carregando...',
      editTitleTooltip: 'Editar título',
      labelTitle: 'Título',
      btnSave: '💾 Salvar',
      btnCancel: 'Cancelar',
      btnDownload: '⬇️ Baixar',
      btnGenerateThumbnail: '🖼️ Gerar thumbnail',
      btnDeleteThumbnail: '🗑️ Deletar thumbnail',
      labelType: 'Tipo',
      labelFormat: 'Formato',
      labelId: 'ID',
      typeVideo: 'Vídeo',
      typeAudio: 'Áudio',
      typeImage: 'Imagem',
      typeOther: 'Arquivo',
      typeMedia: 'Mídia',
      btnDownloadVideo: '⬇️ Baixar vídeo',
      btnDownloadAudio: '⬇️ Baixar áudio',
      btnDownloadImage: '⬇️ Baixar imagem',
      btnDownloadFile: '⬇️ Baixar arquivo',
      videoNotSupported: 'Seu navegador não suporta o elemento de vídeo.',
      audioNotSupported: 'Seu navegador não suporta o elemento de áudio.',
      errorLoadingMediaTitle: 'Erro ao carregar',
      errorRenewStream: 'Não foi possível renovar o link de streaming. Tente recarregar a página.',
      errorMediaRetries: 'Não foi possível carregar a mídia após várias tentativas. Atualize a página.',
      errorDownloadFile: 'Erro ao baixar arquivo: ',
      btnGeneratingThumb: '⏳ Gerando...',
      btnThumbGenerated: '✅ Gerado',
      errorGenerateThumb: 'Erro ao gerar thumbnail: ',
      confirmDeleteThumb: 'Tem certeza que deseja deletar o thumbnail desta mídia?',
      btnDeletingThumb: '⏳ Deletando...',
      btnThumbDeleted: '✅ Deletado',
      errorDeleteThumb: 'Erro ao deletar thumbnail: ',
      errorTitleEmpty: 'O título não pode ficar vazio.',
      btnSaving: '⏳ Salvando...',
      btnSaved: '✅ Salvo',
      errorUpdateMedia: 'Erro ao atualizar mídia: ',
      errorServerResponse: 'Erro na resposta do servidor:',
      confirmDeleteMedia: 'Tem certeza que deseja excluir esta mídia?',
      errorDelete: 'Erro ao excluir',
      errorDeleteWithMsg: 'Erro ao excluir: '
    },
    en: {
      tagline: 'Access your files',
      tabLogin: 'Login',
      tabRegister: 'Register',
      labelCode: 'Code',
      labelPassword: 'Password',
      labelName: 'Name',
      placeholderCode: 'e.g. abc123',
      placeholderName: 'Your full name',
      btnLogin: 'Log In',
      btnLoginLoading: 'Logging in...',
      btnRegister: 'Create Account',
      btnRegisterLoading: 'Creating...',
      footerNoAccount: "Don't have an account?",
      footerSignUp: 'Sign up',
      footerHasAccount: 'Already have an account?',
      footerSignIn: 'Log in',
      msgLoginSuccess: '✅ Login successful!',
      msgLoginErrorDefault: 'Incorrect code or password',
      msgRegisterSuccess: '✅ Account created! Please log in.',
      msgRegisterErrorDefault: 'Error creating account',
      msgConnectionError: '❌ Connection error: ',
      uploadPageTitle: 'Upload - Wireless Drive',
      mediaPageTitle: 'Media - Wireless Drive',
      uploadTitle: 'Upload',
      uploadSubtitle: 'Drag files here or select them to upload (multiple files supported)',
      dropText: 'Drag your files here',
      dropSubtext: 'or',
      selectFilesButton: 'Select files',
      filesSelected: 'Selected files',
      uploadButton: 'Upload files 🚀',
      removeFile: 'Remove',
      tipsTitle: 'Tips',
      selectAtLeastOneFile: '❌ Select at least one file',
      needLoginToUpload: 'You need to sign in to upload files. Redirecting...',
      uploading: 'Uploading...',
      uploadSupportedFormats: 'Supported formats: JPG, PNG, GIF, MP4, WebM, PDF, DOC, DOCX, PPT, XLS, TXT, ZIP',
      uploadImageMaxSize: 'Recommended max size for images: 50MB',
      uploadAutoTitle: 'Titles will be auto-filled from the file name and can be edited later',
      uploadMultipleFiles: 'You can upload multiple files at once',
      generatingThumbnail: 'Generating thumbnail 🖼️',
      uploadedSuccess: '✅ Uploaded successfully',
      uploadSuccessAll: '✅ All {count} file(s) were uploaded successfully!',
      uploadPartialSuccess: '⚠️ {uploaded} file(s) uploaded successfully, but {failed} failed.',
      uploadFailedAll: '❌ Failed to upload {count} file(s).',
      sessionExpired: '❌ Session expired',
      errorUpload: '❌ Error: {error}',
      errorConnection: '❌ Error: connection failure',
      unknownError: 'Unknown error',

      // Dashboard
      navUpload: '📤 Upload',
      navLogout: '🚪 Log Out',
      searchPlaceholder: 'Search media...',
      searchClearAria: 'Clear search',
      toastCloseAria: 'Close',
      btnGenerateThumbnails: '🖼️ Generate Thumbnails',
      btnGeneratingThumbnails: 'Generating...',
      viewToggleToCompact: 'View as Compact Grid',
      viewToggleToGrid: 'View as Grid',
      sortByDate: 'Date',
      sortAsc: 'A→Z',
      sortDesc: 'Z→A',
      noResultsFor: 'No media found for',
      noMediaYet: 'No media yet. Do an',
      uploadCta: 'upload',
      loadMediaError: 'Error loading media.',
      lightboxLoading: 'Loading media...',
      lightboxLoadingSub: 'Please wait while the media is transferred',
      retryBtn: '↻ Try Again',
      closeBtn: 'Close',
      errorTitle: 'Loading Error',
      errorVideo: 'Could not load the video. Please try again.',
      errorMedia: 'Could not load the media. Check your internet connection and try again.',
      allThumbnailsExist: 'All media already have thumbnails!',
      errorGeneratingThumbnailFor: 'Error generating thumbnail for',
      errorGeneratingThumbnails: 'Unable to generate thumbnails, server unavailable or unexpected error',
      authForbidden: "You don't have permission to access this. Redirecting to login...",
      authExpired: 'Your session has expired. Redirecting to login...',

      // Media page
      navShuffle: 'Random',
      backLink: '← Back',
      deleteLink: '🗑️ Delete',
      loadingTitle: 'Loading...',
      loadingDate: '📅 Loading...',
      editTitleTooltip: 'Edit title',
      labelTitle: 'Title',
      btnSave: '💾 Save',
      btnCancel: 'Cancel',
      btnDownload: '⬇️ Download',
      btnGenerateThumbnail: '🖼️ Generate thumbnail',
      btnDeleteThumbnail: '🗑️ Delete thumbnail',
      labelType: 'Type',
      labelFormat: 'Format',
      labelId: 'ID',
      typeVideo: 'Video',
      typeAudio: 'Audio',
      typeImage: 'Image',
      typeOther: 'File',
      typeMedia: 'Media',
      btnDownloadVideo: '⬇️ Download video',
      btnDownloadAudio: '⬇️ Download audio',
      btnDownloadImage: '⬇️ Download image',
      btnDownloadFile: '⬇️ Download file',
      videoNotSupported: 'Your browser does not support the video element.',
      audioNotSupported: 'Your browser does not support the audio element.',
      errorLoadingMediaTitle: 'Error loading',
      errorRenewStream: 'Could not renew the streaming link. Please reload the page.',
      errorMediaRetries: 'Could not load the media after several attempts. Please refresh the page.',
      errorDownloadFile: 'Error downloading file: ',
      btnGeneratingThumb: '⏳ Generating...',
      btnThumbGenerated: '✅ Generated',
      errorGenerateThumb: 'Error generating thumbnail: ',
      confirmDeleteThumb: "Are you sure you want to delete this media's thumbnail?",
      btnDeletingThumb: '⏳ Deleting...',
      btnThumbDeleted: '✅ Deleted',
      errorDeleteThumb: 'Error deleting thumbnail: ',
      errorTitleEmpty: 'Title cannot be empty.',
      btnSaving: '⏳ Saving...',
      btnSaved: '✅ Saved',
      errorUpdateMedia: 'Error updating media: ',
      errorServerResponse: 'Server response error:',
      confirmDeleteMedia: 'Are you sure you want to delete this media?',
      errorDelete: 'Error deleting',
      errorDeleteWithMsg: 'Error deleting: '

    }
  };

  function detectLang() {
    const saved = localStorage.getItem('lang');
    if (saved && translations[saved]) return saved;
    return (navigator.language || '').toLowerCase().startsWith('pt') ? 'pt' : 'en';
  }

  let currentLang = detectLang();

  function t(key, vars) {
    var text = translations[currentLang][key] ?? translations.pt[key] ?? key;
    if (!vars || typeof vars !== 'object') {
      return text;
    }
    return text.replace(/\{([^}]+)\}/g, function (match, name) {
      return vars[name] !== undefined ? vars[name] : match;
    });
  }

  function applyLang() {
    document.documentElement.lang = currentLang === 'pt' ? 'pt-BR' : 'en';

    document.querySelectorAll('[data-i18n]').forEach(function (el) {
      const key = el.dataset.i18n;
      if (translations[currentLang][key] !== undefined) {
        el.textContent = translations[currentLang][key];
      }
    });

    document.querySelectorAll('[data-i18n-placeholder]').forEach(function (el) {
      const key = el.dataset.i18nPlaceholder;
      if (translations[currentLang][key] !== undefined) {
        el.placeholder = translations[currentLang][key];
      }
    });

    document.querySelectorAll('[data-i18n-title]').forEach(function (el) {
      const key = el.dataset.i18nTitle;
      if (translations[currentLang][key] !== undefined) {
        el.setAttribute('title', translations[currentLang][key]);
      }
    });

    document.querySelectorAll('[data-i18n-aria]').forEach(function (el) {
      const key = el.dataset.i18nAria;
      if (translations[currentLang][key] !== undefined) {
        el.setAttribute('aria-label', translations[currentLang][key]);
      }
    });

    document.querySelectorAll('[data-lang-option]').forEach(function (el) {
      el.classList.toggle('active', el.dataset.langOption === currentLang);
    });
  }

  function setLang(lang) {
    if (!translations[lang]) return;
    currentLang = lang;
    localStorage.setItem('lang', lang);
    applyLang();
    document.dispatchEvent(new CustomEvent('i18n:change', { detail: { lang: currentLang } }));
  }

  window.i18n = {
    t: t,
    setLang: setLang,
    get lang() { return currentLang; },
    get locale() { return currentLang === 'pt' ? 'pt-BR' : 'en-US'; }
  };

  document.addEventListener('DOMContentLoaded', applyLang);
})();