#version 330

uniform sampler2D tex;
uniform float threshold;

in vec2 tex_coord;
out vec4 color;

void main()
{
    vec4 in_color = texture2D(tex, tex_coord);
    float brightness = dot(in_color.rgb, vec3(0.2126, 0.7152, 0.0722));

    if (brightness > threshold) {
        color = in_color;
    } else {
        color = vec4(0.0);
    }
}