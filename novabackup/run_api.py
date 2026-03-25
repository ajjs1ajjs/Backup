def main():
    try:
        from novabackup.api import get_app
        import uvicorn

        app = get_app()
        uvicorn.run(app, host="0.0.0.0", port=8000)
    except Exception as e:
        print(f"Failed to start API: {e}")


if __name__ == "__main__":
    main()
