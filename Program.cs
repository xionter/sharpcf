using System.Diagnostics;
using System.IO;

namespace Maker;
class Program
{
    public static async Task Main()
    {
        var parser = new Process();
        parser.StartInfo.FileName = "go";
        parser.StartInfo.Arguments = "run .";
        parser.StartInfo.WorkingDirectory = "parser";
        parser.StartInfo.UseShellExecute = false;
        parser.Start();
        parser.WaitForExit();
        
        var contestPath = File.ReadAllText(Path.Join("parser", "current_contest.txt")).Trim();
        var problemDirs = Directory.GetDirectories(contestPath);
        
        var testTemplate = Path.Join("templates", "tests_template.txt");
        var programTemplate = Path.Join("templates", "program_template.txt");
        var tasks = problemDirs
            .Select(dir => CreateProject(dir, testTemplate, programTemplate));
        await Task.WhenAll(tasks);
    }
    public static Task CreateProject(string path, string testTemplate, string programTemplate)
    {
        return Task.Run(() => {
                var problem = Path.GetFileName(path);
                var proc = Process.Start("dotnet", $"new console -n {problem} -o {path}");
                proc?.WaitForExit();
                File.WriteAllText($"{path}/Program.cs", File.ReadAllText(programTemplate));

                var testsPath = Path.Join(path, "tests.sh");
                File.WriteAllText(testsPath, File.ReadAllText(testTemplate));
                Process.Start("chmod", $"+x {testsPath}");
                }
                );
    }
}
